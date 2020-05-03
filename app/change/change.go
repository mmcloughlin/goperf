// Package change implements change detection in benchmark timeseries.
package change

import (
	"github.com/mmcloughlin/cb/pkg/units"
)

type Change struct {
	CommitIndex int
	EffectSize  float64
	Pre         Stats
	Post        Stats
}

func (c *Change) Delta() float64 {
	return c.Post.Mean - c.Pre.Mean
}

func (c *Change) Percent() float64 {
	return 100 * c.Delta() / c.Pre.Mean
}

// Type of a change: improvement or regression.
type Type int

// Supported change types.
const (
	TypeUnknown Type = iota
	TypeUnchanged
	TypeImprovement
	TypeRegression
)

//go:generate enumer -type Type -output type_enum.go -trimprefix Type -transform snake

// Classify a change from pre to post in the given unit.
func Classify(pre, post float64, unit string) Type {
	if post == pre {
		return TypeUnchanged
	}

	d := units.ImprovementDirectionForUnit(unit)
	if d != units.ImprovementDirectionSmaller && d != units.ImprovementDirectionLarger {
		return TypeUnknown
	}

	delta := post - pre
	if d == units.ImprovementDirectionSmaller {
		delta = -delta
	}

	if delta > 0 {
		return TypeImprovement
	}
	return TypeRegression
}
