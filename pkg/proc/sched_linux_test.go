package proc

import (
	"reflect"
	"testing"

	"golang.org/x/sys/unix"
)

func TestAffinitySelf(t *testing.T) {
	cpus, err := AffinitySelf()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cpus)
	if len(cpus) == 0 {
		t.Fatal("empty set")
	}
}

func TestCPUList(t *testing.T) {
	expect := []int{3, 7, 11}

	s := &unix.CPUSet{}
	s.Zero()
	for _, cpu := range expect {
		s.Set(cpu)
	}

	if got := cpulist(s); !reflect.DeepEqual(expect, got) {
		t.Fatalf("got list %v; expect %v", got, expect)
	}
}
