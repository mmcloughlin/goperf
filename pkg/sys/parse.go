package sys

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mmcloughlin/cb/pkg/cfg"
)

type fileproperty struct {
	filename string
	alias    string
	parser   func(string) (cfg.Value, error)
	doc      string
	tags     []cfg.Tag
}

func property(filename string, parser func(string) (cfg.Value, error), doc string, tags ...cfg.Tag) fileproperty {
	return fileproperty{
		filename: filename,
		parser:   parser,
		doc:      doc,
		tags:     tags,
	}
}

func perfproperty(filename string, parser func(string) (cfg.Value, error), doc string) fileproperty {
	return property(filename, parser, doc, cfg.TagPerfCritical)
}

func (p fileproperty) key() cfg.Key {
	if p.alias != "" {
		return cfg.Key(p.alias)
	}
	return cfg.Key(strings.ReplaceAll(p.filename, "_", ""))
}

func (p fileproperty) parse(s string) (cfg.Entry, error) {
	v, err := p.parser(s)
	if err != nil {
		return nil, err
	}
	return cfg.Property(p.key(), p.doc, v, p.tags...), nil
}

func parsefiles(root string, properties []fileproperty) (cfg.Configuration, error) {
	c := cfg.Configuration{}
	for _, p := range properties {
		filename := filepath.Join(root, p.filename)
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		data := strings.TrimSpace(string(b))
		prop, err := p.parse(data)
		if err != nil {
			return nil, err
		}
		c = append(c, prop)
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
