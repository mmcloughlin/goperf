package coordinator

import (
	"errors"
	"io"
	"regexp"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/entity"
	"github.com/mmcloughlin/cb/pkg/job"
)

type JobsRequest struct {
	Worker string
}

func (r *JobsRequest) Validate() error {
	return validateWorker(r.Worker)
}

type JobsResponse struct {
	Jobs []*Job `json:"jobs"`
}

func NoJobsAvailable() *JobsResponse {
	return &JobsResponse{
		Jobs: []*Job{},
	}
}

type Job struct {
	UUID      uuid.UUID `json:"uuid"`
	CommitSHA string    `json:"commit_sha"`
	Suite     job.Suite `json:"suite"`
}

type StatusChangeRequest struct {
	Worker string
	UUID   uuid.UUID
	From   []entity.TaskStatus
	To     entity.TaskStatus
}

func (r *StatusChangeRequest) Validate() error {
	return validateWorker(r.Worker)
}

type ResultRequest struct {
	io.Reader // data file

	Worker string
	UUID   uuid.UUID
}

func (r *ResultRequest) Validate() error {
	return validateWorker(r.Worker)
}

var workerRegexp = regexp.MustCompile(`^[a-z][a-z0-9\-]*$`)

func validateWorker(worker string) error {
	if worker == "" {
		return errors.New("empty worker name")
	}

	if match := workerRegexp.FindString(worker); match == "" {
		return errors.New("worker must contain only lowercase letters, numbers and hyphens, and start with a letter")
	}

	return nil
}
