package peakhour

import "time"

var (
	EndMidnight = &Time{
		Hour:   23,
		Minute: 59,
	}

	StartMidnight = &Time{
		Hour:   0,
		Minute: 0,
	}
)

type Time struct {
	Hour   int
	Minute int
}

func NewTimeNow() *Time {
	return NewTime(time.Now())
}

func NewTime(t time.Time) *Time {
	return &Time{
		Hour:   t.Hour(),
		Minute: t.Minute(),
	}
}

func (t *Time) IsGreaterThanOrEqual(t1 *Time) bool {
	if t.Hour == t1.Hour {
		return t.Minute >= t1.Minute
	}

	return t.Hour > t1.Hour
}

func (t *Time) IsLessThanOrEqual(t1 *Time) bool {
	if t.Hour == t1.Hour {
		return t.Minute <= t1.Minute
	}

	return t.Hour < t1.Hour
}

func (t *Time) IsGreaterThan(t1 *Time) bool {
	if t.Hour == t1.Hour {
		return t.Minute > t1.Minute
	}

	return t.Hour > t1.Hour
}

func (t *Time) IsLessThan(t1 *Time) bool {
	if t.Hour == t1.Hour {
		return t.Minute < t1.Minute
	}

	return t.Hour < t1.Hour
}

// t subtracted by t1, will always return 0 <= result < 24 hour
func (t *Time) Subtract(t1 *Time) (time.Duration, bool) {
	minuteSrc := t.Minute
	hourSrc := t.Hour
	isNextDay := false

	minuteResult := 0
	hourResult := 0

	if minuteSrc < t1.Minute {
		minuteSrc += 60
		hourSrc -= 1
	}

	minuteResult = minuteSrc - t1.Minute

	if hourSrc < t1.Hour {
		hourSrc += 24
		isNextDay = true
	}

	hourResult = hourSrc - t1.Hour
	return time.Duration(minuteResult+hourResult*60) * time.Minute, isNextDay
}
