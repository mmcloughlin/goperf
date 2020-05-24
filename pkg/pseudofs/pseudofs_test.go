package pseudofs

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/internal/test"
)

func TestInts(t *testing.T) {
	cases := []struct {
		Name   string
		Data   string
		Expect []int
	}{
		// Empty file edge case.
		{
			Name:   "empty",
			Data:   "",
			Expect: nil,
		},
		// Multi-line case like /sys/fs/cgroup/cpuset/tasks.
		{
			Name:   "multiline",
			Data:   "123\n456\n789\n",
			Expect: []int{123, 456, 789},
		},
		// Single-line case like /sys/devices/system/cpu/cpu0/cpufreq/scaling_available_frequencies.
		{
			Name:   "singleline",
			Data:   "600000 1500000 \n",
			Expect: []int{600000, 1500000},
		},
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.Name, func(t *testing.T) {
			d := test.TempDir(t)
			path := filepath.Join(d, "ints")
			err := ioutil.WriteFile(path, []byte(c.Data), 0o644)
			if err != nil {
				t.Fatal(err)
			}

			got, err := Ints(path)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(c.Expect, got); diff != "" {
				t.Fatalf("mismatch\n%s", diff)
			}
		})
	}
}
