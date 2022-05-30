package async

import "time"

// TempoPolicy is a Policy that triggers a rebuld at Tempo.
type TempoPolicy struct {
	tempo time.Duration
}

func NewTempoPolicy(tempo time.Duration) *TempoPolicy {
	return &TempoPolicy{
		tempo: tempo,
	}
}

func (t *TempoPolicy) C() <-chan time.Time {
	return time.NewTicker(t.tempo).C
}

func (t *TempoPolicy) Mutation(*mutation) {
	// This Policy doesn't care about mutation traffic.
}
