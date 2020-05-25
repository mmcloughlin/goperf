// Package pseudofs provides helpers for interacting with pseudo filesystems such as procfs and sysfs.
package pseudofs

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/mmcloughlin/goperf/internal/errutil"
)

// String reads a string from path, trimming whitespace.
func String(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return strings.TrimSpace(string(b)), nil
}

// WriteString writes a string to a file, followed by a newline.
func WriteString(path, s string) error {
	return WriteFile(path, []byte(s+"\n"), 0o644)
}

// Int reads an integer from a file.
func Int(path string) (int, error) {
	s, err := String(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(s)
}

// WriteInt writes an integer to a file.
func WriteInt(path string, n int) error {
	return WriteString(path, strconv.Itoa(n)+"\n")
}

// Flag reads a boolean from a file, represented as 0 or 1.
func Flag(path string) (bool, error) {
	s, err := String(path)
	if err != nil {
		return false, err
	}

	switch s {
	case "1":
		return true, nil
	case "0":
		return false, nil
	default:
		return false, errutil.AssertionFailure("unexpected file contents %q", s)
	}
}

// WriteFlag writes a boolean to a file.
func WriteFlag(path string, enabled bool) error {
	data := "0\n"
	if enabled {
		data = "1\n"
	}
	return WriteFile(path, []byte(data), 0o644)
}

// Ints reads a list of integers from a file.
func Ints(path string) (_ []int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer errutil.CheckClose(&err, f)

	var ns []int
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanWords)
	for s.Scan() {
		n, err := strconv.Atoi(s.Text())
		if err != nil {
			return nil, err
		}
		ns = append(ns, n)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return ns, nil
}

// WriteInts writes a list of integers to a file with one write per entry.
func WriteInts(path string, ns []int) error {
	// Note the warning in cpuset(7):
	//
	// Warning: only one PID may be written to the tasks file at a
	// time.  If a string is written that contains more than one PID,
	// only the first one will be used.
	for _, n := range ns {
		if err := WriteInt(path, n); err != nil {
			return err
		}
	}
	return nil
}

// WriteFile writes data to path with a single write syscall.
func WriteFile(path string, data []byte, perm uint32) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("write %s: %w", path, err)
		}
	}()

	// Open.
	mode := unix.O_WRONLY | unix.O_CREAT | unix.O_TRUNC
	fd, err := unix.Open(path, mode, perm)
	if err != nil {
		return err
	}

	defer func() {
		if errc := unix.Close(fd); errc != nil && err == nil {
			err = errc
		}
	}()

	// Write.
	n, err := unix.Write(fd, data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}

	return nil
}
