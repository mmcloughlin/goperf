package job

import "encoding/json"

type Job struct {
	Toolchain Toolchain `json:"toolchain"`
	Suites    []Suite   `json:"suite"`
}

type Toolchain struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

type Suite struct {
	Module Module `json:"module`
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
