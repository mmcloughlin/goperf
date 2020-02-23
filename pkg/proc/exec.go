package proc

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// Exec executes the given command using the execve(2) system call, inheriting
// environment.
func Exec(args []string) error {
	if len(args) == 0 {
		return errors.New("no command provided")
	}
	argv0, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return syscall.Exec(argv0, args, os.Environ())
}
