package entity

import (
	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/change"
)

// ChangeSummary is a change with associated metadata.
type ChangeSummary struct {
	Benchmark       *Benchmark
	EnvironmentUUID uuid.UUID

	CommitSHA     string
	CommitSubject string

	change.Change
}
