package peakhour

type Period struct {
	Start *Time
	End   *Time
}

func (p *Period) IsTimeInPeriod(t *Time) bool {
	if p.Start.IsLessThanOrEqual(t) && p.End.IsGreaterThanOrEqual(t) {
		return true
	}

	return false
}
