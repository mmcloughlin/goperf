package cpuset

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

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

func listfile(path string) (Set, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	s := strings.TrimSpace(string(b))
	return ParseList(s)
}
