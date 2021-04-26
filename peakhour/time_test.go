package peakhour

import (
	"testing"
	"time"
)

func TestTime_IsGreaterThan(t *testing.T) {
	tests := map[string]struct {
		T1       *Time
		T2       *Time
		Expected bool
	}{
		"greater than same hour": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 30},
			Expected: true,
		},
		"greater than different hour": {
			T1:       &Time{11, 40},
			T2:       &Time{10, 30},
			Expected: true,
		},
		"less than": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 41},
			Expected: false,
		},
		"equal": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 40},
			Expected: false,
		},
		"edge case": {
			T1:       StartMidnight,
			T2:       EndMidnight,
			Expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.T1.IsGreaterThan(tc.T2) != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, tc.T1.IsGreaterThan(tc.T2))
			}
		})
	}
}

func TestTime_IsGreaterThanOrEqual(t *testing.T) {
	tests := map[string]struct {
		T1       *Time
		T2       *Time
		Expected bool
	}{
		"greater than same hour": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 30},
			Expected: true,
		},
		"greater than different hour": {
			T1:       &Time{11, 40},
			T2:       &Time{10, 30},
			Expected: true,
		},
		"less than": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 41},
			Expected: false,
		},
		"equal": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 40},
			Expected: true,
		},
		"edge case": {
			T1:       StartMidnight,
			T2:       EndMidnight,
			Expected: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.T1.IsGreaterThanOrEqual(tc.T2) != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, tc.T1.IsGreaterThan(tc.T2))
			}
		})
	}
}

func TestTime_IsLessThan(t *testing.T) {
	tests := map[string]struct {
		T1       *Time
		T2       *Time
		Expected bool
	}{
		"less than same hour": {
			T1:       &Time{10, 20},
			T2:       &Time{10, 30},
			Expected: true,
		},
		"less than different hour": {
			T1:       &Time{9, 40},
			T2:       &Time{10, 30},
			Expected: true,
		},
		"greater than": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 39},
			Expected: false,
		},
		"equal": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 40},
			Expected: false,
		},
		"edge case": {
			T1:       StartMidnight,
			T2:       EndMidnight,
			Expected: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.T1.IsLessThan(tc.T2) != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, tc.T1.IsGreaterThan(tc.T2))
			}
		})
	}
}

func TestTime_IsLessThanOrEqual(t *testing.T) {
	tests := map[string]struct {
		T1       *Time
		T2       *Time
		Expected bool
	}{
		"less than same hour": {
			T1:       &Time{10, 20},
			T2:       &Time{10, 30},
			Expected: true,
		},
		"less than different hour": {
			T1:       &Time{9, 40},
			T2:       &Time{10, 30},
			Expected: true,
		},
		"greater than": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 39},
			Expected: false,
		},
		"equal": {
			T1:       &Time{10, 40},
			T2:       &Time{10, 40},
			Expected: true,
		},
		"edge case": {
			T1:       StartMidnight,
			T2:       EndMidnight,
			Expected: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.T1.IsLessThanOrEqual(tc.T2) != tc.Expected {
				t.Errorf("expected %v, got %v", tc.Expected, tc.T1.IsGreaterThan(tc.T2))
			}
		})
	}
}

func TestTime_Subtract(t *testing.T) {
	tests := map[string]struct {
		T1               *Time
		T2               *Time
		ExpectedDuration time.Duration
		ExpectedNexDay   bool
	}{
		"normal case": {
			T1:               &Time{10, 30},
			T2:               &Time{10, 20},
			ExpectedDuration: 10 * time.Minute,
			ExpectedNexDay:   false,
		},
		"normal case different hour": {
			T1:               &Time{10, 30},
			T2:               &Time{9, 20},
			ExpectedDuration: 70 * time.Minute,
			ExpectedNexDay:   false,
		},
		"normal case minute less than": {
			T1:               &Time{10, 20},
			T2:               &Time{9, 30},
			ExpectedDuration: 50 * time.Minute,
			ExpectedNexDay:   false,
		},
		"equal": {
			T1:               &Time{10, 20},
			T2:               &Time{10, 20},
			ExpectedDuration: 0 * time.Minute,
			ExpectedNexDay:   false,
		},
		"next day case": {
			T1:               &Time{10, 20},
			T2:               &Time{10, 30},
			ExpectedDuration: 23*time.Hour + 50*time.Minute,
			ExpectedNexDay:   true,
		},
		"next day case different hour": {
			T1:               &Time{9, 20},
			T2:               &Time{10, 30},
			ExpectedDuration: 22*time.Hour + 50*time.Minute,
			ExpectedNexDay:   true,
		},
		"next day case edge case": {
			T1:               StartMidnight,
			T2:               EndMidnight,
			ExpectedDuration: 1 * time.Minute,
			ExpectedNexDay:   true,
		},
		"normal case edge case": {
			T1:               EndMidnight,
			T2:               StartMidnight,
			ExpectedDuration: 23*time.Hour + 59*time.Minute,
			ExpectedNexDay:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d, b := tc.T1.Subtract(tc.T2)
			if d != tc.ExpectedDuration {
				t.Errorf("duration expected %v, got %v", tc.ExpectedDuration, d)
			}

			if b != tc.ExpectedNexDay {
				t.Errorf("is next day expected %v, got %v", tc.ExpectedNexDay, b)
			}
		})
	}
}

func TestNewTimeNow(t *testing.T) {
	expectedTime := &Time{10, 29}
	Now = func() time.Time {
		return time.Date(1, 1, 1, expectedTime.Hour, expectedTime.Minute, 0, 0, time.Now().Location())
	}

	t1 := NewTimeNow()

	if *t1 != *expectedTime {
		t.Errorf("expected %v, got %v", *t1, *expectedTime)
	}
}
