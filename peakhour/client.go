package peakhour

import (
	"fmt"
	"strings"
	"time"
)

type Client struct {
	Periods    []Period
	IsMidnight bool
}

func NewPeakHour(periods []string) *Client {
	p, i := ParseRangeString(periods)
	return &Client{
		Periods:    p,
		IsMidnight: i,
	}
}

func ParseRangeString(periodsStr []string) ([]Period, bool) {
	periods := make([]Period, 0)
	for _, periodStr := range periodsStr {
		p := strings.Split(periodStr, "-")
		if len(p) != 2 {
			panic(fmt.Sprintf("invalid peak hour ranges: %s", periodStr))
		}

		start, err := time.Parse("15:04", p[0])
		if err != nil {
			panic(err)
		}

		end, err := time.Parse("15:04", p[1])
		if err != nil {
			panic(err)
		}

		if start.After(end) {
			periodStart := Period{
				Start: NewTime(start),
				End:   EndMidnight,
			}

			periodEnd := Period{
				Start: StartMidnight,
				End:   NewTime(end),
			}

			periods = MergePeriod(periods, periodStart)
			periods = MergePeriod(periods, periodEnd)
		} else {
			periods = MergePeriod(periods, Period{
				Start: NewTime(start),
				End:   NewTime(end),
			})
		}
	}

	// check is midnight
	isMidnight := false
	isStartMidnight := false
	isEndMidnight := false
	for _, period := range periods {
		if period.Start == StartMidnight {
			isStartMidnight = true
		}

		if period.End == EndMidnight {
			isEndMidnight = true
		}
	}

	if isStartMidnight && isEndMidnight {
		isMidnight = true
	}

	return periods, isMidnight
}

func MergePeriod(periods []Period, addedPeriod Period) []Period {
	newPeriods := make([]Period, 0)
	intersectedPeriod := make([]Period, 0)
	for _, period := range periods {
		if period.Start.IsGreaterThan(addedPeriod.End) {
			newPeriods = append(newPeriods, period)
			continue
		}

		if period.End.IsLessThan(addedPeriod.Start) {
			newPeriods = append(newPeriods, period)
			continue
		}

		intersectedPeriod = append(intersectedPeriod, period)
	}

	for _, period := range intersectedPeriod {
		if period.Start.IsLessThan(addedPeriod.Start) {
			addedPeriod.Start = period.Start
		}

		if period.End.IsGreaterThan(addedPeriod.End) {
			addedPeriod.End = period.End
		}
	}

	newPeriods = append(newPeriods, addedPeriod)

	return newPeriods
}

func (c *Client) IsPeakHourNow() bool {
	now := NewTimeNow()

	for _, period := range c.Periods {
		if period.IsTimeInPeriod(now) {
			return true
		}
	}

	return false
}

func (c *Client) GetNearestEndPeakHour() time.Time {
	result := &Time{}
	minD := 24 * time.Hour
	tNow := time.Now()
	now := NewTime(tNow)
	isNextDay := false
	for _, period := range c.Periods {
		if c.IsMidnight {
			if period.End == EndMidnight {
				continue
			}
		}

		d, next := period.End.Subtract(now)

		if minD > d {
			minD = d
			isNextDay = next
			result = period.Start
		}
	}

	if isNextDay {
		tNow.Add(24 * time.Hour)
	}

	return time.Date(tNow.Year(), tNow.Month(), tNow.Day(), result.Hour, result.Minute, 0, 0, tNow.Location())
}

func (c *Client) GetNearestStartPeakHour() time.Time {
	result := &Time{}
	minD := 24 * time.Hour
	tNow := time.Now()
	now := NewTime(tNow)
	isNextDay := false
	for _, period := range c.Periods {
		if c.IsMidnight {
			if period.End == StartMidnight {
				continue
			}
		}

		d, next := period.Start.Subtract(now)

		if minD > d {
			minD = d
			isNextDay = next
			result = period.Start
		}
	}

	if isNextDay {
		tNow.Add(24 * time.Hour)
	}

	return time.Date(tNow.Year(), tNow.Month(), tNow.Day(), result.Hour, result.Minute, 0, 0, tNow.Location())
}
