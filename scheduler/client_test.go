package scheduler

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"preemptible-lifecycle-scheduler/cluster"
	"preemptible-lifecycle-scheduler/peakhour"
	"testing"
	"time"
)

func TestClient_GetPeakHourState(t *testing.T) {
	tests := map[string]struct {
		PeriodStr   []string
		CurrentTime *peakhour.Time
		Expected    string
	}{
		"in peak hour normal case": {
			PeriodStr:   []string{"10:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 12, Minute: 31},
			Expected:    InPeakHour,
		},
		"in peak hour equal start": {
			PeriodStr:   []string{"10:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 10, Minute: 00},
			Expected:    InPeakHour,
		},
		"in peak hour equal end": {
			PeriodStr:   []string{"10:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 15, Minute: 00},
			Expected:    OutsidePeakHour,
		},
		"start peak hour": {
			PeriodStr:   []string{"10:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 9, Minute: 50},
			Expected:    StartPeakHour,
		},
		"start peak hour, midnight case": {
			PeriodStr:   []string{"00:05-15:00"},
			CurrentTime: &peakhour.Time{Hour: 23, Minute: 55},
			Expected:    StartPeakHour,
		},
		"start peak hour, edge case": {
			PeriodStr:   []string{"00:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 23, Minute: 55},
			Expected:    StartPeakHour,
		},
		"in peak hour, end midnight case": {
			PeriodStr:   []string{"22:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 23, Minute: 59},
			Expected:    OutsidePeakHour,
		},
		"in peak hour, start midnight case": {
			PeriodStr:   []string{"22:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 00, Minute: 00},
			Expected:    InPeakHour,
		},
		"outside peak hour, near end": {
			PeriodStr:   []string{"22:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 15, Minute: 01},
			Expected:    OutsidePeakHour,
		},
		"outside peak hour, normal case": {
			PeriodStr:   []string{"22:00-15:00"},
			CurrentTime: &peakhour.Time{Hour: 18, Minute: 00},
			Expected:    OutsidePeakHour,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			peakhour.Now = func() time.Time {
				return time.Date(1, 1, 1, tc.CurrentTime.Hour, tc.CurrentTime.Minute, 0, 0, time.Now().Location())
			}

			ph, err := peakhour.NewClient(tc.PeriodStr)
			if err != nil {
				t.Errorf("failed to create peak hour client %v", err)
			}

			client := NewClient(nil, ph, 15)
			if client.GetPeakHourState() != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, client.GetPeakHourState())
			}
		})
	}
}

func TestClient_CalculateNextSchedule(t *testing.T) {
	tests := map[string]struct {
		PeriodStr     []string
		NodeCreatedTs []time.Time
		CurrentTime   time.Time
		Expected      time.Duration
	}{
		"next peak hour": {
			PeriodStr: []string{"11:00-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 10, 22, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 1, 10, 23, 0, 0, time.Now().Location()),
			Expected:    7 * time.Minute,
		},
		"one node": {
			PeriodStr: []string{"12:00-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 11, 30, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 10, 23, 0, 0, time.Now().Location()),
			Expected:    37 * time.Minute,
		},
		"many nodes": {
			PeriodStr: []string{"12:00-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 11, 30, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 12, 45, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 10, 21, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 00, 00, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 10, 23, 0, 0, time.Now().Location()),
			Expected:    37 * time.Minute,
		},
		"next period midnight": {
			PeriodStr: []string{"00:35-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 11, 30, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 12, 45, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 21, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 57, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
			Expected:    6 * time.Minute,
		},
		"node midnight": {
			PeriodStr: []string{"12:00-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 11, 30, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 12, 45, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 21, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 00, 35, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
			Expected:    6 * time.Minute,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			peakhour.Now = func() time.Time {
				return tc.CurrentTime
			}

			ph, err := peakhour.NewClient(tc.PeriodStr)
			if err != nil {
				t.Errorf("failed to create peak hour client %v", err)
			}

			nodes := make([]corev1.Node, 0)
			for _, ts := range tc.NodeCreatedTs {
				nodes = append(nodes, corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						CreationTimestamp: v1.Time{Time: ts},
					},
				})
			}

			client := NewClient(&cluster.Client{}, ph, 15)
			if client.CalculateNextSchedule(nodes) != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, client.CalculateNextSchedule(nodes))
			}
		})
	}
}

type MockClusterClient struct {
	ProcessedTs []time.Time
}

func NewMockClusterClient() *MockClusterClient {
	return &MockClusterClient{
		ProcessedTs: make([]time.Time, 0),
	}
}

func (c *MockClusterClient) GetPreemptibleNodes() (*corev1.NodeList, error) {
	return nil, nil
}

func (c *MockClusterClient) ProcessNode(node corev1.Node) (err error) {
	c.ProcessedTs = append(c.ProcessedTs, c.GetNodeCreatedTime(node))
	return nil
}

func (c *MockClusterClient) GetNodeCreatedTime(node corev1.Node) time.Time {
	cc := &cluster.Client{}
	return cc.GetNodeCreatedTime(node)
}

func TestClient_ProcessNodesStartPeakHour(t *testing.T) {
	tests := map[string]struct {
		PeriodStr     []string
		NodeCreatedTs []time.Time
		CurrentTime   time.Time
		Expected      []time.Time
	}{
		"normal case": {
			PeriodStr: []string{"09:00-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 10, 22, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 8, 15, 0, 0, time.Now().Location()),
			Expected: []time.Time{
				time.Date(1, 1, 1, 10, 22, 0, 0, time.Now().Location()),
			},
		},
		"normal case, empty": {
			PeriodStr: []string{"09:00-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 15, 01, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 8, 15, 0, 0, time.Now().Location()),
			Expected:    []time.Time{},
		},
		"equal": {
			PeriodStr: []string{"09:00-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 15, 00, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 8, 15, 0, 0, time.Now().Location()),
			Expected: []time.Time{
				time.Date(1, 1, 1, 15, 00, 0, 0, time.Now().Location()),
			},
		},
		"more nodes": {
			PeriodStr: []string{"09:00-15:00"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 15, 00, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 9, 00, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 8, 50, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 50, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 8, 30, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 0, 0, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 17, 29, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 8, 45, 0, 0, time.Now().Location()),
			Expected: []time.Time{
				time.Date(1, 1, 1, 15, 00, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 9, 00, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 8, 50, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 50, 0, 0, time.Now().Location()),
			},
		},
		"mid night": {
			PeriodStr: []string{"00:00-04:00", "23:55-23:59"},
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 2, 15, 00, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 56, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 00, 00, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 4, 01, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 2, 35, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 23, 45, 0, 0, time.Now().Location()),
			Expected: []time.Time{
				time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 00, 00, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 2, 35, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 56, 0, 0, time.Now().Location()),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			peakhour.Now = func() time.Time {
				return tc.CurrentTime
			}

			ph, err := peakhour.NewClient(tc.PeriodStr)
			if err != nil {
				t.Errorf("failed to create peak hour client %v", err)
			}

			nodes := make([]corev1.Node, 0)
			for _, ts := range tc.NodeCreatedTs {
				nodes = append(nodes, corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						CreationTimestamp: v1.Time{Time: ts},
					},
				})
			}

			cc := NewMockClusterClient()
			client := NewClient(cc, ph, 15)
			client.ProcessNodesStartPeakHour(nodes)

			if !isTimestampsEqual(cc.ProcessedTs, tc.Expected) {
				t.Errorf("expected %v, got %v", tc.Expected, cc.ProcessedTs)
			}
		})
	}
}

func isTimestampsEqual(t1 []time.Time, t2 []time.Time) bool {
	if len(t1) != len(t2) {
		return false
	}

	exist := make(map[time.Time]struct{}, 0)
	for _, t := range t1 {
		exist[t] = struct{}{}
	}

	for _, t := range t2 {
		if _, ok := exist[t]; !ok {
			return false
		}
	}

	return true
}

func TestClient_ProcessNodesOutsidePeakHour(t *testing.T) {
	tests := map[string]struct {
		NodeCreatedTs       []time.Time
		CurrentTime         time.Time
		ProcessedExpected   []time.Time
		UnprocessedExpected []time.Time
	}{
		"normal case": {
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 10, 22, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 10, 15, 0, 0, time.Now().Location()),
			ProcessedExpected: []time.Time{
				time.Date(1, 1, 1, 10, 22, 0, 0, time.Now().Location()),
			},
			UnprocessedExpected: []time.Time{},
		},
		"many nodes": {
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 10, 22, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 45, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 15, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 46, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 10, 14, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 0, 0, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 10, 15, 0, 0, time.Now().Location()),
			ProcessedExpected: []time.Time{
				time.Date(1, 1, 1, 10, 22, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 45, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 10, 15, 0, 0, time.Now().Location()),
			},
			UnprocessedExpected: []time.Time{
				time.Date(1, 1, 1, 10, 46, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 10, 14, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 0, 0, 0, 0, time.Now().Location()),
			},
		},
		"mid night": {
			NodeCreatedTs: []time.Time{
				time.Date(1, 1, 1, 23, 58, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 0, 0, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 0, 5, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 23, 50, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 1, 50, 0, 0, time.Now().Location()),
			},
			CurrentTime: time.Date(1, 1, 2, 23, 55, 0, 0, time.Now().Location()),
			ProcessedExpected: []time.Time{
				time.Date(1, 1, 1, 23, 58, 0, 0, time.Now().Location()),
				time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 0, 0, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 0, 5, 0, 0, time.Now().Location()),
			},
			UnprocessedExpected: []time.Time{
				time.Date(1, 1, 2, 23, 50, 0, 0, time.Now().Location()),
				time.Date(1, 1, 2, 1, 50, 0, 0, time.Now().Location()),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			peakhour.Now = func() time.Time {
				return tc.CurrentTime
			}

			ph, err := peakhour.NewClient([]string{})
			if err != nil {
				t.Errorf("failed to create peak hour client %v", err)
			}

			nodes := make([]corev1.Node, 0)
			for _, ts := range tc.NodeCreatedTs {
				nodes = append(nodes, corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						CreationTimestamp: v1.Time{Time: ts},
					},
				})
			}

			cc := NewMockClusterClient()
			client := NewClient(cc, ph, 15)

			unprocessedNodes := client.ProcessNodesOutsidePeakHour(nodes)
			unprocessedTs := make([]time.Time, 0)
			for _, node := range unprocessedNodes {
				unprocessedTs = append(unprocessedTs, cc.GetNodeCreatedTime(node))
			}

			if !isTimestampsEqual(cc.ProcessedTs, tc.ProcessedExpected) {
				t.Errorf("processed timstamp expected %v, got %v", tc.ProcessedExpected, cc.ProcessedTs)
			}

			if !isTimestampsEqual(unprocessedTs, tc.UnprocessedExpected) {
				t.Errorf("unprocessed timestamp expected %v, got %v", tc.UnprocessedExpected, unprocessedTs)
			}
		})
	}
}
