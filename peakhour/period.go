package peakhour

// Start is always less than or equal to end
type Period struct {
	Start *Time
	End   *Time
}

func (p *Period) IsTimeInPeriod(t *Time) bool {
	if p.Start.IsLessThanOrEqual(t) && p.End.IsGreaterThan(t) {
		return true
	}

	return false
}
