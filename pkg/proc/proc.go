// Package proc implements process manipulation through syscalls.
package proc

import "golang.org/x/sys/unix"

// Writable checks whether path exists and the calling process can write to it.
func Writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

// Readable checks whether path exists and the calling process can read it.
func Readable(path string) bool {
	return unix.Access(path, unix.R_OK) == nil
}
