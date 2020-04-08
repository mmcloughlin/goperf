package job

import (
	"encoding/json"
	"time"
)

type Job struct {
	Toolchain Toolchain `json:"toolchain"`
	Suites    []Suite   `json:"suite"`
}

type Toolchain struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

type Suite struct {
	Module     Module        `json:"module"`
	Tests      string        `json:"tests,omitempty"`
	Short      bool          `json:"short,omitempty"`
	Benchmarks string        `json:"benchmarks,omitempty"`
	BenchTime  time.Duration `json:"benchtime_ns,omitempty"`
	Timeout    time.Duration `json:"timeout_ns,omitempty"`
}

// TestRegex returns the regular expression controlling which tests are run.
func (s *Suite) TestRegex() string {
	if s.Tests == "" {
		return "."
	}
	return s.Tests
}

// BenchmarkRegex returns the regular expression controlling which benchmarks are run.
func (s *Suite) BenchmarkRegex() string {
	if s.Benchmarks == "" {
		return "."
	}
	return s.Benchmarks
}

// BenchmarkTime returns the minimum amount of time each benchmark is run for.
func (s *Suite) BenchmarkTime() time.Duration {
	if s.BenchTime == 0 {
		return time.Second
	}
	return s.BenchTime
}

type Module struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

func (m Module) String() string {
	s := m.Path
	if m.Version != "" {
		s += "@" + m.Version
	}
	return s
}

func Marshal(j *Job) ([]byte, error) {
	return json.Marshal(j)
}

func Unmarshal(b []byte) (*Job, error) {
	j := &Job{}
	if err := json.Unmarshal(b, j); err != nil {
		return nil, err
	}
	return j, nil
}
