package sched

import (
	"time"

	"github.com/mmcloughlin/goperf/app/db"
)

// NewDefault builds a scheduler with sensible defaults.
func NewDefault(d *db.DB) Scheduler {
	// Recent commits.
	pri := TimeSinceSmoothStep(
		60*24*time.Hour, PriorityHigh,
		365*24*time.Hour, PriorityIdle,
	)
	recent := NewRecentCommits(d, pri)

	// Retries.
	retries := NewRetry(d, 5, time.Hour)

	return CompositeScheduler(
		recent,
		retries,
	)
}
