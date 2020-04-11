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
	TaskStatusCreated             TaskStatus = iota + 1 // initial state
	TaskStatusInProgress                                // task has been sent to a worker and is in progress
	TaskStatusResultUploadStarted                       // result upload has begun
	TaskStatusResultUploaded                            // result upload complete
	TaskStatusCompleteSuccess                           // completed successfully
	TaskStatusCompleteError                             // completed with error
	TaskStatusHalted                                    // worker stopped processing the task
)

//go:generate enumer -type TaskStatus -output taskstatus_enum.go -trimprefix TaskStatus -transform snake

// IsComplete reports whether this task is finished.
func (s TaskStatus) IsComplete() bool {
	return s == TaskStatusCompleteSuccess || s == TaskStatusCompleteError
}

// IsPending reports whether this task is in a pending state.
func (s TaskStatus) IsPending() bool {
	return !s.IsComplete()
}

// TaskStatusCompleteValues returns all complete task states.
func TaskStatusCompleteValues() []TaskStatus { return filterTaskStatusValues(TaskStatus.IsComplete) }

// TaskStatusPendingValues returns all pending task states.
func TaskStatusPendingValues() []TaskStatus { return filterTaskStatusValues(TaskStatus.IsPending) }

func filterTaskStatusValues(predicate func(TaskStatus) bool) []TaskStatus {
	filtered := []TaskStatus{}
	for _, status := range TaskStatusValues() {
		if predicate(status) {
			filtered = append(filtered, status)
		}
	}
	return filtered
}

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
