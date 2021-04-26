package peakhour

import "testing"

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
