package peakhour

import "testing"

func TestPeriod_IsTimeInPeriod(t *testing.T) {
	tests := map[string]struct {
		Period   *Period
		T        *Time
		Expected bool
	}{
		"normal case": {
			Period: &Period{
				Start: &Time{9, 10},
				End:   &Time{10, 20},
			},
			T:        &Time{9, 20},
			Expected: true,
		},
		"normal case, same hour": {
			Period: &Period{
				Start: &Time{9, 10},
				End:   &Time{9, 20},
			},
			T:        &Time{9, 15},
			Expected: true,
		},
		"equal": {
			Period: &Period{
				Start: &Time{9, 10},
				End:   &Time{9, 20},
			},
			T:        &Time{9, 10},
			Expected: true,
		},
		"less than start": {
			Period: &Period{
				Start: &Time{9, 10},
				End:   &Time{9, 20},
			},
			T:        &Time{8, 5},
			Expected: false,
		},
		"greater than start": {
			Period: &Period{
				Start: &Time{9, 10},
				End:   &Time{9, 20},
			},
			T:        &Time{10, 5},
			Expected: false,
		},
		"edge case": {
			Period: &Period{
				Start: StartMidnight,
				End:   &Time{9, 20},
			},
			T:        StartMidnight,
			Expected: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.Period.IsTimeInPeriod(tc.T) != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, tc.Period.IsTimeInPeriod(tc.T))
			}
		})
	}
}
