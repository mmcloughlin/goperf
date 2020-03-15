package cpuset

import (
	"reflect"
	"testing"
)

func TestCPUSetTasks(t *testing.T) {
	cpuset := NewCPUSetPath("testdata/cpuset")
	tasks, err := cpuset.Tasks()
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1154 {
		t.Fail()
	}
	head := []int{1, 2, 4, 6, 7}
	if !reflect.DeepEqual(head, tasks[:5]) {
		t.Fail()
	}
}

func TestCPUSetFlags(t *testing.T) {
	cpuset := NewCPUSetPath("testdata/cpuset")
	cases := []struct {
		Flag   func(*CPUSet) (bool, error)
		Expect bool
	}{
		{(*CPUSet).NotifyOnRelease, false},
		{(*CPUSet).CPUExclusive, true},
		{(*CPUSet).MemExclusive, true},
		{(*CPUSet).MemHardwall, false},
		{(*CPUSet).MemoryMigrate, false},
		{(*CPUSet).MemoryPressureEnabled, false},
		{(*CPUSet).MemorySpreadPage, false},
		{(*CPUSet).MemorySpreadSlab, false},
		{(*CPUSet).SchedLoadBalance, true},
	}
	for _, c := range cases {
		got, err := c.Flag(cpuset)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.Expect {
			t.Fail()
		}
	}
}

func TestCPUSetLists(t *testing.T) {
	cpuset := NewCPUSetPath("testdata/cpuset")
	cases := []struct {
		Set    func(*CPUSet) (Set, error)
		Expect string
	}{
		{(*CPUSet).CPUs, "0-23"},
		{(*CPUSet).Mems, "0"},
	}
	for _, c := range cases {
		got, err := c.Set(cpuset)
		if err != nil {
			t.Fatal(err)
		}
		if got.FormatList() != c.Expect {
			t.Fail()
		}
	}
}

func TestCPUSetInts(t *testing.T) {
	cpuset := NewCPUSetPath("testdata/cpuset")
	cases := []struct {
		Int    func(*CPUSet) (int, error)
		Expect int
	}{
		{(*CPUSet).MemoryPressure, 0},
		{(*CPUSet).SchedRelaxDomainLevel, -1},
	}
	for _, c := range cases {
		got, err := c.Int(cpuset)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.Expect {
			t.Fail()
		}
	}
}
