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

			for _, node := range nodes.Items {
				createdAt := c.Cluster.GetNodeCreatedTime(node)

				// node is nearly terminated
				if createdAt.Add(24*time.Hour).Sub(peakhour.Now()) <= DefaultGracefulDuration {
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

			sleepDuration := c.PeakHours.GetNearestEndPeakHour().Sub(peakhour.Now())
			log.Printf("waiting for next peak hour period: %s", sleepDuration.String())
			time.Sleep(sleepDuration)
		}
	}
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

	minT = minT.Add(-1 * DefaultGracefulDuration)
	return minT.Sub(peakhour.Now())
}

func (c *Client) GetPeakHourState() string {
	if c.PeakHours.IsPeakHourNow() {
		return InPeakHour
	}

	if c.PeakHours.GetNearestStartPeakHour().Sub(peakhour.Now()) <= DefaultGracefulDuration {
		return StartPeakHour
	}

	return OutsidePeakHour
}
