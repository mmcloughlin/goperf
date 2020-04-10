package coordinator

import (
	"errors"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/pkg/job"
)

type JobsRequest struct {
	Worker string
}

func (r *JobsRequest) Validate() error {
	if r.Worker == "" {
		return errors.New("missing worker name")
	}
	return nil
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

type StartRequest struct {
	Worker string
	UUID   uuid.UUID
}

func (r *StartRequest) Validate() error {
	if r.Worker == "" {
		return errors.New("missing worker name")
	}
	return nil
}

type StartResponse struct {
}
