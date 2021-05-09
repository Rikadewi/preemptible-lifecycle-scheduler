package scheduler

import (
	corev1 "k8s.io/api/core/v1"
	"log"
	"preemptible-lifecycle-scheduler/peakhour"
	"time"
)

const (
	peakHourMultiplier = 2

	InPeakHour      = "in peak hour"
	OutsidePeakHour = "outside peak hour"
	StartPeakHour   = "start peak hour"
)

type ClusterClient interface {
	GetPreemptibleNodes() (*corev1.NodeList, error)
	ProcessNode(node *corev1.Node) (err error)
	GetNodeCreatedTime(node corev1.Node) time.Time
}

type Client struct {
	Cluster        ClusterClient
	PeakHours      *peakhour.Client
	GracefulPeriod time.Duration
}

func NewClient(cluster ClusterClient, peakHour *peakhour.Client, gracefulPeriod int) *Client {
	return &Client{
		Cluster:        cluster,
		PeakHours:      peakHour,
		GracefulPeriod: peakHourMultiplier * time.Duration(gracefulPeriod) * time.Minute,
	}
}

func (c *Client) Start() {
	for {
		currentState := c.GetPeakHourState()
		log.Printf("current state: %s", currentState)

		switch currentState {
		case InPeakHour:
			sleepDuration := c.PeakHours.GetNearestEndPeakHour().Sub(peakhour.Now())
			log.Printf("in peak hour, waiting %s", sleepDuration.String())
			time.Sleep(sleepDuration)

		case OutsidePeakHour:
			nodes, err := c.Cluster.GetPreemptibleNodes()
			if err != nil {
				log.Printf("failed to get preemptible nodes: %v", err)
				break
			}

			if len(nodes.Items) == 0 {
				continue
			}
			log.Printf("%d nodes found", len(nodes.Items))

			unprocessedNodes := c.ProcessNodesOutsidePeakHour(nodes.Items)

			sleepDuration := c.CalculateNextSchedule(unprocessedNodes)
			log.Printf("waiting for next schedule: %s", sleepDuration.String())
			time.Sleep(sleepDuration)

		case StartPeakHour:
			nodes, err := c.Cluster.GetPreemptibleNodes()
			if err != nil {
				log.Printf("failed to get preemptible nodes: %v", err)
				break
			}

			if len(nodes.Items) == 0 {
				continue
			}
			log.Printf("%d nodes found", len(nodes.Items))

			c.ProcessNodesStartPeakHour(nodes.Items)

			sleepDuration := c.PeakHours.GetNearestEndPeakHour().Sub(peakhour.Now())
			log.Printf("waiting for next peak hour period: %s", sleepDuration.String())
			time.Sleep(sleepDuration)
		}
	}
}

func (c *Client) ProcessNodesStartPeakHour(nodes []corev1.Node) {
	for _, node := range nodes {
		createdAt := c.Cluster.GetNodeCreatedTime(node)
		log.Println(createdAt.String())
		endPeakHour := c.PeakHours.GetNearestEndPeakHour()

		// node won't survive next peak hour period
		if endPeakHour.After(createdAt.Add(24*time.Hour)) || endPeakHour.Equal(createdAt.Add(24*time.Hour)) {
			err := c.Cluster.ProcessNode(&node)
			if err != nil {
				log.Printf("failed to process node: %v", err)
				continue
			}
		}
	}
}

func (c *Client) ProcessNodesOutsidePeakHour(nodes []corev1.Node) []corev1.Node {
	unprocessedNodes := make([]corev1.Node, 0)
	for _, node := range nodes {
		createdAt := c.Cluster.GetNodeCreatedTime(node)
		log.Println(createdAt.String())

		// node is nearly terminated
		if createdAt.Add(24*time.Hour).Sub(peakhour.Now()) <= c.GracefulPeriod {
			err := c.Cluster.ProcessNode(&node)
			if err != nil {
				log.Printf("failed to process node: %v", err)
			}
			continue
		}

		unprocessedNodes = append(unprocessedNodes, node)
	}

	return unprocessedNodes
}

func (c *Client) CalculateNextSchedule(nodes []corev1.Node) time.Duration {
	minT := peakhour.Now().Add(24 * time.Hour)
	for _, node := range nodes {
		t := c.Cluster.GetNodeCreatedTime(node).Add(24 * time.Hour)

		if minT.After(t) {
			minT = t
		}
	}

	start := c.PeakHours.GetNearestStartPeakHour()
	if minT.After(start) {
		minT = start
	}

	minT = minT.Add(-1 * c.GracefulPeriod)
	return minT.Sub(peakhour.Now())
}

func (c *Client) GetPeakHourState() string {
	if c.PeakHours.IsPeakHourNow() {
		return InPeakHour
	}

	if c.PeakHours.GetNearestStartPeakHour().Sub(peakhour.Now()) <= c.GracefulPeriod {
		return StartPeakHour
	}

	return OutsidePeakHour
}
