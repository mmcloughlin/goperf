// Code generated by sqlc. DO NOT EDIT.

package db

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TaskStatus string

const (
	TaskStatusCreated             TaskStatus = "created"
	TaskStatusInProgress          TaskStatus = "in_progress"
	TaskStatusCompleteSuccess     TaskStatus = "complete_success"
	TaskStatusCompleteError       TaskStatus = "complete_error"
	TaskStatusResultUploadStarted TaskStatus = "result_upload_started"
	TaskStatusResultUploaded      TaskStatus = "result_uploaded"
	TaskStatusHalted              TaskStatus = "halted"
	TaskStatusStaleTimeout        TaskStatus = "stale_timeout"
)

func (e *TaskStatus) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TaskStatus(s)
	case string:
		*e = TaskStatus(s)
	default:
		return fmt.Errorf("unsupported scan type for TaskStatus: %T", src)
	}
	return nil
}

type TaskType string

const (
	TaskTypeModule TaskType = "module"
)

func (e *TaskType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = TaskType(s)
	case string:
		*e = TaskType(s)
	default:
		return fmt.Errorf("unsupported scan type for TaskType: %T", src)
	}
	return nil
}

type Benchmark struct {
	UUID        uuid.UUID
	PackageUUID uuid.UUID
	FullName    string
	Name        string
	Unit        string
	Parameters  json.RawMessage
}

type Commit struct {
	SHA            []byte
	Tree           []byte
	Parents        pq.ByteaArray
	AuthorName     string
	AuthorEmail    string
	AuthorTime     time.Time
	CommitterName  string
	CommitterEmail string
	CommitTime     time.Time
	Message        string
}

type CommitPosition struct {
	SHA        []byte
	CommitTime time.Time
	Index      int32
}

type CommitRef struct {
	SHA []byte
	Ref string
}

type Datafile struct {
	UUID   uuid.UUID
	Name   string
	SHA256 []byte
}

type Module struct {
	UUID    uuid.UUID
	Path    string
	Version string
}

type Package struct {
	UUID         uuid.UUID
	ModuleUUID   uuid.UUID
	RelativePath string
}

type Property struct {
	UUID   uuid.UUID
	Fields json.RawMessage
}

type Result struct {
	UUID            uuid.UUID
	DatafileUUID    uuid.UUID
	Line            int32
	BenchmarkUUID   uuid.UUID
	CommitSHA       []byte
	EnvironmentUUID uuid.UUID
	MetadataUUID    uuid.UUID
	Iterations      int64
	Value           float64
}

type Task struct {
	UUID             uuid.UUID
	Worker           string
	CommitSHA        []byte
	Type             TaskType
	TargetUUID       uuid.UUID
	Status           TaskStatus
	LastStatusUpdate time.Time
	DatafileUUID     uuid.UUID
}
