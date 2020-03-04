package cpuset

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/mmcloughlin/cb/internal/errutil"
)

// TODO(mbm): consider moving this to a separate package, since there is overlap with the pkg/sys package

func intfile(path string) (int, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(b)))
}

func writeintfile(path string, n int) error {
	data := strconv.Itoa(n) + "\n"
	return writefile(path, []byte(data), 0644)
}

func flagfile(path string) (bool, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return false, err
	}

	switch string(b) {
	case "1\n":
		return true, nil
	case "0\n":
		return false, nil
	default:
		return false, errutil.AssertionFailure("unexpected file contents %q", b)
	}
}

func writeflagfile(path string, enabled bool) error {
	data := "0\n"
	if enabled {
		data = "1\n"
	}
	return writefile(path, []byte(data), 0644)
}

func intsfile(path string) ([]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ns []int
	s := bufio.NewScanner(f)
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

func writeintsfile(path string, ns []int) error {
	// Note the warning in cpuset(7):
	//
	// Warning: only one PID may be written to the tasks file at a
	// time.  If a string is written that contains more than one PID,
	// only the first one will be used.
	for _, n := range ns {
		if err := writeintfile(path, n); err != nil {
			return err
		}
	}
	return nil
}

func listfile(path string) (Set, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := strings.TrimSpace(string(b))
	return ParseList(s)
}

func writelistfile(path string, s Set) error {
	data := s.FormatList() + "\n"
	return writefile(path, []byte(data), 0644)
}

func writefile(path string, data []byte, perm uint32) error {
	// Open.
	mode := unix.O_WRONLY | unix.O_CREAT | unix.O_TRUNC
	fd, err := unix.Open(path, mode, perm)
	if err != nil {
		return err
	}

	// Write.
	n, err := unix.Write(fd, data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}

	// Close.
	if err := unix.Close(fd); err != nil {
		return err
	}

	return nil
}
