package entity

import (
	"time"

	"github.com/google/uuid"
)

// TaskType describes a type of task.
type TaskType uint

// Supported task types.
const (
	TaskTypeModule TaskType = iota + 1 // benchmark a go module
)

//go:generate enumer -type TaskType -output tasktype_enum.go -trimprefix TaskType -transform snake

// TaskStatus describes the state of a task.
type TaskStatus uint

// Supported task status values.
const (
	TaskStatusCreated         TaskStatus = iota + 1 // initial state
	TaskStatusInProgress                            // task has been sent to a worker and is in progress
	TaskStatusCompleteSuccess                       // completed successfully
	TaskStatusCompleteError                         // completed with error
)

//go:generate enumer -type TaskStatus -output taskstatus_enum.go -trimprefix TaskStatus -transform snake

// TaskSpec specifies work required by a task.
type TaskSpec struct {
	Type       TaskType
	TargetUUID uuid.UUID
	CommitSHA  string
}

type Task struct {
	UUID             uuid.UUID
	Worker           string
	Spec             TaskSpec
	Status           TaskStatus
	LastStatusUpdate time.Time
	DatafileUUID     uuid.UUID
}
