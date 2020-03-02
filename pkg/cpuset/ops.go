package cpuset

// MoveTasks moves all tasks from src to dst cpusets.
func MoveTasks(src, dst *CPUSet) error {
	tasks, err := src.Tasks()
	if err != nil {
		return err
	}

	return dst.AddTasks(tasks)
}
