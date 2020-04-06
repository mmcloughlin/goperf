package coordinator

import "github.com/mmcloughlin/cb/pkg/job"

type JobsRequest struct {
	Worker string `json:"worker"`
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
	CommitSHA string
	Suite     job.Suite
}
