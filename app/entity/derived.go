package entity

import (
	"time"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/change"
)

// Point in a benchmark result timeseries.
type Point struct {
	ResultUUID      uuid.UUID
	EnvironmentUUID uuid.UUID
	CommitSHA       string
	CommitTime      time.Time
	Value           float64
}

// Points is a series of measurements.
type Points []*Point

// Values returns the series of point values.
func (p Points) Values() []float64 {
	xs := make([]float64, len(p))
	for i := range p {
		xs[i] = p[i].Value
	}
	return xs
}

// ChangeSummary is a change with associated metadata.
type ChangeSummary struct {
	Benchmark       *Benchmark
	EnvironmentUUID uuid.UUID

	CommitSHA     string
	CommitSubject string

	change.Change
}
