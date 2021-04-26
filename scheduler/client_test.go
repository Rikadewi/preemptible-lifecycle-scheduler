package scheduler

import (
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
			Expected:    InPeakHour,
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
			Expected:    InPeakHour,
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

			client := NewClient(nil, ph)
			if client.GetPeakHourState() != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, client.GetPeakHourState())
			}
		})
	}
}
