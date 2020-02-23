package runner

import "os/exec"

// Wrapper is a method of modifying commands before execution.
type Wrapper func(*exec.Cmd)

// RunUnder builds a wrapper that inserts the named command and arguments before a command before execution.
func RunUnder(name string, arg ...string) Wrapper {
	w := exec.Command(name, arg...)
	return func(cmd *exec.Cmd) {
		args := []string{}
		args = append(args, w.Args...)
		args = append(args, "--")
		args = append(args, argv(cmd)...)

		cmd.Path = w.Path
		cmd.Args = args
	}
}

func argv(cmd *exec.Cmd) []string {
	a := []string{cmd.Path}
	if len(cmd.Args) > 1 {
		a = append(a, cmd.Args[1:]...)
	}
	return a
}
