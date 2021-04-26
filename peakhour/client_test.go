package peakhour

import (
	"testing"
	"time"
)

func TestParseRangeString(t *testing.T) {
	tests := map[string]struct {
		PeriodStr          []string
		ExpectedPeriod     []Period
		ExpectedIsMidnight bool
		ExpectedErrNotNil  bool
	}{
		"one period": {
			PeriodStr: []string{"11:00-12:00"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{11, 00},
					End:   &Time{12, 00},
				},
			},
			ExpectedIsMidnight: false,
			ExpectedErrNotNil:  false,
		},
		"empty period": {
			PeriodStr:          []string{},
			ExpectedPeriod:     []Period{},
			ExpectedIsMidnight: false,
			ExpectedErrNotNil:  false,
		},
		"normal case": {
			PeriodStr: []string{"11:00-12:00", "15:20-15:40", "23:51-23:59"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{11, 00},
					End:   &Time{12, 00},
				},
				{
					Start: &Time{15, 20},
					End:   &Time{15, 40},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
			},
			ExpectedIsMidnight: false,
			ExpectedErrNotNil:  false,
		},
		"midnight case": {
			PeriodStr: []string{"11:00-12:00", "15:20-15:40", "23:51-04:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{11, 00},
					End:   &Time{12, 00},
				},
				{
					Start: &Time{15, 20},
					End:   &Time{15, 40},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{04, 31},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"overlap end first": {
			PeriodStr: []string{"11:00-12:00", "11:20-15:40", "23:51-04:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{11, 00},
					End:   &Time{15, 40},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{04, 31},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"overlap start first": {
			PeriodStr: []string{"11:00-12:00", "10:09-11:10", "23:51-04:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{10, 9},
					End:   &Time{12, 00},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{04, 31},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"overlap start end first": {
			PeriodStr: []string{"11:00-12:00", "10:09-12:10", "23:51-04:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{10, 9},
					End:   &Time{12, 10},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{04, 31},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"overlap start end second": {
			PeriodStr: []string{"11:00-12:00", "11:09-11:15", "23:51-04:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{11, 00},
					End:   &Time{12, 00},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{04, 31},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"overlap equal": {
			PeriodStr: []string{"11:00-12:00", "12:00-13:00", "23:51-04:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{11, 00},
					End:   &Time{13, 00},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{04, 31},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"overlap all": {
			PeriodStr: []string{"11:00-12:00", "12:00-13:00", "13:00-14:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{11, 00},
					End:   &Time{14, 31},
				},
			},
			ExpectedIsMidnight: false,
			ExpectedErrNotNil:  false,
		},
		"overlap midnight all": {
			PeriodStr: []string{"02:30-05:00", "12:00-23:51", "23:51-04:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{12, 00},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{05, 00},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"overlap midnight": {
			PeriodStr: []string{"02:30-04:00", "12:00-13:00", "23:51-04:31"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{12, 00},
					End:   &Time{13, 00},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{04, 31},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"merge midnight": {
			PeriodStr: []string{"00:00-04:00", "12:00-13:00", "23:51-23:59"},
			ExpectedPeriod: []Period{
				{
					Start: &Time{12, 00},
					End:   &Time{13, 00},
				},
				{
					Start: &Time{23, 51},
					End:   &Time{23, 59},
				},
				{
					Start: &Time{00, 00},
					End:   &Time{04, 00},
				},
			},
			ExpectedIsMidnight: true,
			ExpectedErrNotNil:  false,
		},
		"invalid format separator": {
			PeriodStr:          []string{"02:30-04:00-05:00", "12:00-13:00", "23:51-04:31"},
			ExpectedPeriod:     []Period{},
			ExpectedIsMidnight: false,
			ExpectedErrNotNil:  true,
		},
		"invalid format timestamp": {
			PeriodStr:          []string{"02.30-04:00-05:00", "12:00-13:00", "23:51-04:31"},
			ExpectedPeriod:     []Period{},
			ExpectedIsMidnight: false,
			ExpectedErrNotNil:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, b, err := ParseRangeString(tc.PeriodStr)
			if !isEqualPeriods(p, tc.ExpectedPeriod) {
				t.Errorf("period expected %v, got %v", tc.ExpectedPeriod, p)
			}

			if b != tc.ExpectedIsMidnight {
				t.Errorf("expected midnight expected %v, got %v", tc.ExpectedIsMidnight, b)
			}

			if tc.ExpectedErrNotNil && (err == nil) {
				t.Errorf("expected error expected err not nil")
			}

			if !tc.ExpectedErrNotNil && (err != nil) {
				t.Errorf("expected error expected err nil, got %v", err)
			}
		})
	}
}

func isEqualPeriods(p1 []Period, p2 []Period) bool {
	if len(p1) != len(p2) {
		return false
	}

	startEnd := make(map[Time]Time, 0)
	for _, period := range p1 {
		startEnd[*period.Start] = *period.End
	}

	for _, period := range p2 {
		val, ok := startEnd[*period.Start]
		if !ok {
			return false
		}

		if val != *period.End {
			return false
		}
	}

	return true
}

func TestClient_IsPeakHourNow(t *testing.T) {
	tests := map[string]struct {
		PeriodStr   []string
		CurrentTime *Time
		Expected    bool
	}{
		"normal case": {
			PeriodStr:   []string{"10:00-12:00"},
			CurrentTime: &Time{11, 29},
			Expected:    true,
		},
		"equal": {
			PeriodStr:   []string{"11:29-12:00"},
			CurrentTime: &Time{11, 29},
			Expected:    true,
		},
		"before": {
			PeriodStr:   []string{"10:29-12:00"},
			CurrentTime: &Time{10, 20},
			Expected:    false,
		},
		"after": {
			PeriodStr:   []string{"10:29-12:00"},
			CurrentTime: &Time{12, 20},
			Expected:    false,
		},
		"before midnight": {
			PeriodStr:   []string{"23:29-12:00"},
			CurrentTime: &Time{23, 35},
			Expected:    true,
		},
		"after midnight": {
			PeriodStr:   []string{"23:29-12:00"},
			CurrentTime: &Time{11, 35},
			Expected:    true,
		},
		"edge case": {
			PeriodStr:   []string{"23:29-12:00"},
			CurrentTime: StartMidnight,
			Expected:    true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			Now = func() time.Time {
				return time.Date(1, 1, 1, tc.CurrentTime.Hour, tc.CurrentTime.Minute, 0, 0, time.Now().Location())
			}

			client, err := NewClient(tc.PeriodStr)
			if err != nil {
				t.Errorf("failed to create client %v", err)
			}

			if client.IsPeakHourNow() != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, client.IsPeakHourNow())
			}
		})
	}
}

func TestClient_GetNearestEndPeakHour(t *testing.T) {
	tests := map[string]struct {
		PeriodStr   []string
		CurrentTime time.Time
		Expected    time.Time
	}{
		"one period": {
			PeriodStr:   []string{"10:00-12:00"},
			CurrentTime: time.Date(1, 1, 1, 10, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 12, 00, 0, 0, time.Now().Location()),
		},
		"normal case": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-05:00"},
			CurrentTime: time.Date(1, 1, 1, 10, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 12, 00, 0, 0, time.Now().Location()),
		},
		"before midnight case": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-04:00"},
			CurrentTime: time.Date(1, 1, 1, 22, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 2, 04, 00, 0, 0, time.Now().Location()),
		},
		"after midnight case": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-04:31"},
			CurrentTime: time.Date(1, 1, 1, 03, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 05, 00, 0, 0, time.Now().Location()),
		},
		"equal": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-04:31"},
			CurrentTime: time.Date(1, 1, 1, 12, 00, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 12, 00, 0, 0, time.Now().Location()),
		},
		"end midnight": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-23:59"},
			CurrentTime: time.Date(1, 1, 1, 12, 01, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 23, 59, 0, 0, time.Now().Location()),
		},
		"start midnight": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "00:00-04:50"},
			CurrentTime: time.Date(1, 1, 1, 12, 01, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 2, 05, 00, 0, 0, time.Now().Location()),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			Now = func() time.Time {
				return tc.CurrentTime
			}

			client, err := NewClient(tc.PeriodStr)
			if err != nil {
				t.Errorf("failed to create client %v", err)
			}

			if client.GetNearestEndPeakHour() != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, client.GetNearestEndPeakHour())
			}
		})
	}
}

func TestClient_GetNearestStartPeakHour(t *testing.T) {
	tests := map[string]struct {
		PeriodStr   []string
		CurrentTime time.Time
		Expected    time.Time
	}{
		"one period": {
			PeriodStr:   []string{"10:00-12:00"},
			CurrentTime: time.Date(1, 1, 1, 10, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 2, 10, 00, 0, 0, time.Now().Location()),
		},
		"normal case": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-05:00"},
			CurrentTime: time.Date(1, 1, 1, 10, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 22, 00, 0, 0, time.Now().Location()),
		},
		"before midnight case": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-04:00"},
			CurrentTime: time.Date(1, 1, 1, 22, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 2, 04, 31, 0, 0, time.Now().Location()),
		},
		"after midnight case": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-04:31"},
			CurrentTime: time.Date(1, 1, 1, 03, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 10, 00, 0, 0, time.Now().Location()),
		},
		"before midnight case merge time": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-04:31"},
			CurrentTime: time.Date(1, 1, 1, 22, 23, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 2, 10, 00, 0, 0, time.Now().Location()),
		},
		"equal": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-04:31"},
			CurrentTime: time.Date(1, 1, 1, 10, 00, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 10, 00, 0, 0, time.Now().Location()),
		},
		"end midnight": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "22:00-23:59"},
			CurrentTime: time.Date(1, 1, 1, 12, 01, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 1, 22, 00, 0, 0, time.Now().Location()),
		},
		"next day": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00"},
			CurrentTime: time.Date(1, 1, 1, 12, 01, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 2, 04, 31, 0, 0, time.Now().Location()),
		},
		"start midnight": {
			PeriodStr:   []string{"04:31-05:00", "10:00-12:00", "00:00-04:50"},
			CurrentTime: time.Date(1, 1, 1, 12, 01, 0, 0, time.Now().Location()),
			Expected:    time.Date(1, 1, 2, 00, 00, 0, 0, time.Now().Location()),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			Now = func() time.Time {
				return tc.CurrentTime
			}

			client, err := NewClient(tc.PeriodStr)
			if err != nil {
				t.Errorf("failed to create client %v", err)
			}

			if client.GetNearestStartPeakHour() != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, client.GetNearestStartPeakHour())
			}
		})
	}
}
