package entity

import (
	"time"

	"github.com/google/uuid"
)

// Point in a benchmark result timeseries.
type Point struct {
	ResultUUID      uuid.UUID
	EnvironmentUUID uuid.UUID
	CommitSHA       string
	CommitTime      time.Time
	Value           float64
}
