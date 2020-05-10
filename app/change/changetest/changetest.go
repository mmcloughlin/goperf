// Package changetest provides utilities for change detection testing.
package changetest

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/mmcloughlin/cb/app/trace"
)

// Case is a change detection test case.
type Case struct {
	Expect []int        `json:"expect"`
	Series trace.Series `json:"series"`
}

// ReadCaseFile reads a test case from filename.
func ReadCaseFile(filename string) (*Case, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := new(Case)
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}

	return c, nil
}

// WriteCaseFile writes the test case c to filename.
func WriteCaseFile(filename string, c *Case) error {
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, b, 0o644)
}

// WriteNewCaseFile writes a new test case with the "expect" field empty.
func WriteNewCaseFile(filename string, series trace.Series) error {
	return WriteCaseFile(filename, &Case{
		Expect: []int{},
		Series: series,
	})
}

// Filename returns a recommended filename for the given trace ID.
func Filename(id trace.ID) string {
	return strings.ReplaceAll(id.String(), "/", "_") + ".json"
}
