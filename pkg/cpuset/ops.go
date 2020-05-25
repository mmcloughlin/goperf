package cpuset

import (
	"errors"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/mmcloughlin/goperf/internal/errutil"
)

// MoveResult records the outcome of a batch task migration from one cpuset to
// another.
type MoveResult struct {
	// Tasks successfully moved.
	Moved []int
	// Tasks which could not be found.
	Nonexistent []int
	// Invalid tasks were unmovable for some reason. For example, "bound" tasks
	// that are already pinned to CPUs cannot be moved, nor can the "kthreadd"
	// task.
	Invalid []int
}

// MoveTasks attempts to move all tasks from src to dst cpusets, returning the
// result in a MoveResult struct. Certain error cases are common, for example
// tasks bound to CPUs cannot be moved, and it's also possible for tasks to
// terminate while the move is taking place. These cases will not cause the
// entire operation to error, rather they will be recorded in the MoveResult
// return value for inspection by the caller. Other error cases will be bubbled
// up as errors from MoveTasks.
func MoveTasks(src, dst *CPUSet) (*MoveResult, error) {
	// Fetch tasks from the source.
	tasks, err := src.Tasks()
	if err != nil {
		return nil, err
	}

	// Attempt to move each one.
	r := &MoveResult{}
	for _, task := range tasks {
		err := dst.AddTask(task)

		// Successful move.
		if err == nil {
			r.Moved = append(r.Moved, task)
			continue
		}

		// Inspect errno.
		var errno syscall.Errno
		if !errors.As(err, &errno) {
			return nil, errutil.AssertionFailure("error %q should be errno", err)
		}

		switch errno {
		case unix.EINVAL:
			// Reference: https://github.com/torvalds/linux/blob/531ca1b1f0f5a873d89950d9683f44d9f7001cd3/kernel/cgroup.c#L2274-L2283
			//
			//		/*
			//		 * Workqueue threads may acquire PF_THREAD_BOUND and become
			//		 * trapped in a cpuset, or RT worker may be born in a cgroup
			//		 * with no rt_runtime allocated.  Just say no.
			//		 */
			//		if (tsk == kthreadd_task || (tsk->flags & PF_THREAD_BOUND)) {
			//			ret = -EINVAL;
			//			rcu_read_unlock();
			//			goto out_unlock_cgroup;
			//		}
			//
			r.Invalid = append(r.Invalid, task)

		case unix.ESRCH:
			// Reference: https://github.com/torvalds/linux/blob/531ca1b1f0f5a873d89950d9683f44d9f7001cd3/kernel/cgroup.c#L2241-L2246
			//
			//			tsk = find_task_by_vpid(pid);
			//			if (!tsk) {
			//				rcu_read_unlock();
			//				ret= -ESRCH;
			//				goto out_unlock_cgroup;
			//			}
			//
			r.Nonexistent = append(r.Nonexistent, task)

		default:
			return nil, errno
		}
	}

	return r, nil
}
