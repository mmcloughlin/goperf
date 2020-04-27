package change

import "math"

type Change struct {
	CommitIndex int
	EffectSize  float64
	Pre         Stats
	Post        Stats
}

type Stats struct {
	N        int
	Mean     float64
	Variance float64
}

func (s Stats) Stddev() float64 { return math.Sqrt(s.Variance) }
