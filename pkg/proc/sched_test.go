package proc

import "testing"

func TestPolicyString(t *testing.T) {
	cases := []struct {
		Policy Policy
		Expect string
	}{
		{SCHED_OTHER, "SCHED_OTHER"},
		{SCHED_FIFO, "SCHED_FIFO"},
		{SCHED_RR, "SCHED_RR"},
		{SCHED_BATCH, "SCHED_BATCH"},
		{SCHED_IDLE, "SCHED_IDLE"},
		{SCHED_DEADLINE, "SCHED_DEADLINE"},
		{42, "42"},
	}
	for _, c := range cases {
		if got := c.Policy.String(); got != c.Expect {
			t.Errorf("%d.String() = %q; expect %q", c.Policy, got, c.Expect)
		}
	}
}
