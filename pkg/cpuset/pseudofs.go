package cpuset

import (
	"io/ioutil"
	"strings"

	"github.com/mmcloughlin/cb/pkg/pseudofs"
)

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
	return pseudofs.WriteFile(path, []byte(data), 0o644)
}
