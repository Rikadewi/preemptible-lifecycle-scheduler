package scheduler

import (
	corev1 "k8s.io/api/core/v1"
	"log"
	"preemptible-lifecycle-scheduler/cluster"
	"preemptible-lifecycle-scheduler/peakhour"
	"time"
)

const (
	DefaultGracefulDuration = 30 * time.Minute

	InPeakHour      = "in peak hour"
	OutsidePeakHour = "outside peak hour"
	StartPeakHour   = "start peak hour"
)

type Client struct {
	Cluster   *cluster.Client
	PeakHours *peakhour.Client
}

func NewClient(cluster *cluster.Client, peakHour *peakhour.Client) *Client {
	return &Client{
		Cluster:   cluster,
		PeakHours: peakHour,
	}
}

func (c *Client) Start() {
	for {
		currentState := c.GetPeakHourState()
		switch currentState {
		case InPeakHour:
			log.Println("in peak hour, waiting...")
			sleepDuration := c.PeakHours.GetNearestEndPeakHour().Sub(time.Now())
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

			for _, node := range nodes.Items {
				createdAt := c.Cluster.GetNodeCreatedTime(node)

				// node is nearly terminated
				if createdAt.Add(24*time.Hour).Sub(time.Now()) < DefaultGracefulDuration {
					err := c.Cluster.ProcessNode(node)
					if err != nil {
						log.Printf("failed to delete node: %v", err)
						continue
					}
				}
			}

			log.Println("waiting for next schedule")
			sleepDuration := c.CalculateNextSchedule(nodes.Items)
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

			for _, node := range nodes.Items {
				createdAt := c.Cluster.GetNodeCreatedTime(node)
				endPeakHour := c.PeakHours.GetNearestEndPeakHour()

				// node won't survive next peak hour period
				if endPeakHour.After(createdAt.Add(24 * time.Hour)) {
					err := c.Cluster.ProcessNode(node)
					if err != nil {
						log.Printf("failed to delete node: %v", err)
						continue
					}
				}
			}

			log.Println("waiting for next peak hour period")

			sleepDuration := c.PeakHours.GetNearestEndPeakHour().Sub(time.Now())
			time.Sleep(sleepDuration)
		}

	}
}

func (c *Client) CalculateNextSchedule(nodes []corev1.Node) time.Duration {
	minT := time.Now().Add(24 * time.Hour)
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

	return minT.Sub(time.Now())
}

func (c *Client) GetPeakHourState() string {
	if c.PeakHours.IsPeakHourNow() {
		return InPeakHour
	}

	if c.PeakHours.GetNearestStartPeakHour().Sub(time.Now()) < DefaultGracefulDuration {
		return StartPeakHour
	}

	return OutsidePeakHour
}
