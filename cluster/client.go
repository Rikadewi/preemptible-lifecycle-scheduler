package cluster

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"preemptible-lifecycle-scheduler/config"
	"time"
)

var (
	CheckPodInterval       = 10 * time.Second
	ProcessingNodeInterval = 1 * time.Minute
)

type Client struct {
	KubeClient    *kubernetes.Clientset
	DeleteTimeout time.Duration
}

func NewClient(cfg *config.Config) (*Client, error) {
	var kubernetesConfig *rest.Config
	var err error
	if cfg.Environment == config.EnvDevelopment {
		kubernetesConfig, err = clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	} else {
		kubernetesConfig, err = clientcmd.BuildConfigFromFlags("", "")
	}
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubernetesConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		KubeClient:    clientset,
		DeleteTimeout: time.Duration(cfg.GracefulPeriod) * time.Minute,
	}, nil
}

func (c *Client) GetPreemptibleNodes() (*corev1.NodeList, error) {
	log.Printf("scanning nodes")
	nodes, err := c.KubeClient.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: "cloud.google.com/gke-preemptible=true",
	})
	if err != nil {
		return nil, err
	}

	// filter out exception node
	items := make([]corev1.Node, 0)
	for _, node := range nodes.Items {
		if val, ok := node.Labels["cloud.google.com/gke-nodepool"]; ok {
			if val == "prem-test-pool" {
				continue
			}
		}

		items = append(items, node)
	}

	nodes.Items = items
	return nodes, err
}

func (c *Client) ProcessNode(node corev1.Node) (err error) {
	log.Printf("processing node %s", node.Name)

	doneProcessing := make(chan bool)
	go func() {
		for {
			err = c.UnScheduleNode(node)
			if err != nil {
				log.Printf("error unschedule node: %s, err :%v", node.Name, err)
				time.Sleep(ProcessingNodeInterval)
				continue
			}

			err = c.DeletePods(node)
			if err != nil {
				log.Printf("error delete pods: %s, err :%v", node.Name, err)
				time.Sleep(ProcessingNodeInterval)
				continue
			}

			err = c.DeleteNode(node)
			if err != nil {
				log.Printf("error delete node: %s, err :%v", node.Name, err)
				time.Sleep(ProcessingNodeInterval)
				continue
			}

			doneProcessing <- true
		}
	}()

	select {
	case <-doneProcessing:
		log.Println("done processing node")
		break
	case <-time.After(c.DeleteTimeout):
		log.Println("timeout processing node")
		return nil
	}

	return nil
}

func (c *Client) UnScheduleNode(node corev1.Node) error {
	node.Spec.Unschedulable = true
	_, err := c.KubeClient.CoreV1().Nodes().Update(&node)
	return err
}

func (c *Client) DeletePods(node corev1.Node) error {
	pods, err := c.GetPods(node)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		// TODO: try to check the grace period in delete option
		err = c.KubeClient.CoreV1().Pods(pod.Namespace).Delete(pod.Name, &metav1.DeleteOptions{})
		if err != nil {
			log.Printf("failed to delete pod %s: %v", pod.Name, err)
			continue
		}
	}

	doneDeleting := make(chan bool)
	go func() {
		// check whether all pods have been terminated
		for {
			pods, err := c.GetPods(node)
			if err != nil {
				log.Printf("error get pods from node: %s, err: %v", node.Name, err)
				time.Sleep(CheckPodInterval)
				continue
			}

			if len(pods) == 0 {
				doneDeleting <- true
				return
			}

			// wait for pod to be deleted
			time.Sleep(CheckPodInterval)
		}
	}()

	select {
	case <-doneDeleting:
		log.Println("done deleting")
		break
	case <-time.After(c.DeleteTimeout):
		log.Println("timeout deleting node")
		return nil
	}

	return nil
}

// Get all application pods, filtered out pod from kube-system namespace and DaemonSet.
func (c *Client) GetPods(node corev1.Node) (pods []corev1.Pod, err error) {
	pods = make([]corev1.Pod, 0)
	podList, err := c.KubeClient.CoreV1().Pods("").List(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s,metadata.namespace!=kube-system", node.Name),
	})
	if err != nil {
		return
	}

	for _, pod := range podList.Items {
		// filter out pod from DaemonSet
		isDaemonSet := false
		for _, owner := range pod.ObjectMeta.OwnerReferences {
			if owner.Kind == "DaemonSet" {
				isDaemonSet = true
				break
			}
		}

		if isDaemonSet {
			continue
		}

		pods = append(pods, pod)
	}

	return
}

func (c *Client) DeleteNode(node corev1.Node) error {
	// TODO: try to check the grace period in delete option
	log.Printf("deleting node %s", node.Name)
	return c.KubeClient.CoreV1().Nodes().Delete(node.Name, &metav1.DeleteOptions{})
}

func (c *Client) GetNodeCreatedTime(node corev1.Node) time.Time {
	return node.CreationTimestamp.Time
}
