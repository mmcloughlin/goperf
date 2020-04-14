// Package proc implements process manipulation through syscalls.
package proc

import "golang.org/x/sys/unix"

// Writable checks whether the calling process can write to path.
func Writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}
