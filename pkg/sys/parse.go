package sys

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

type fileproperty struct {
	Filename string
	Parser   func(string) (cfg.Value, error)
	Doc      string
}

func parsefiles(root string, properties []fileproperty) (cfg.Configuration, error) {
	c := cfg.Configuration{}
	for _, p := range properties {
		filename := filepath.Join(root, p.Filename)
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		data := strings.TrimSpace(string(b))
		v, err := p.Parser(data)
		if err != nil {
			return nil, err
		}
		c = append(c, cfg.Property(
			cfg.Key(strings.ReplaceAll(p.Filename, "_", "")),
			p.Doc,
			v,
		))
	}
	return c, nil
}

func parseint(s string) (cfg.Value, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return cfg.IntValue(n), nil
}

func parsekhz(s string) (cfg.Value, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return cfg.FrequencyValue(n * 1000), nil
}

func parsebool(s string) (cfg.Value, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return nil, err
	}
	return cfg.BoolValue(b), nil
}

func parsestring(s string) (cfg.Value, error) {
	return cfg.StringValue(s), nil
}

func parsesize(s string) (cfg.Value, error) {
	if len(s) == 0 || s[len(s)-1] != 'K' {
		return nil, errors.New("expected last character of size to be K")
	}
	b, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return nil, err
	}
	return cfg.BytesValue(b * 1024), nil
}
