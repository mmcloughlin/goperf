package coordinator

import (
	"errors"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/pkg/job"
)

type JobsRequest struct {
	Worker string `json:"worker"`
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
	TaskUUID  uuid.UUID
	CommitSHA string
	Suite     job.Suite
}
