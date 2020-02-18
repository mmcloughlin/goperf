package cfg

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// Value is a configuration value.
type Value interface {
	String() string
}

type StringValue string

func (s StringValue) String() string { return string(s) }

type BytesValue uint64

func (b BytesValue) String() string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	i := 0
	x := float64(b)
	for x >= 1024 && i+1 < len(units) {
		x /= 1024
		i++
	}
	return formatfloat(x, 2) + " " + units[i]
}

type PercentageValue float64

func (p PercentageValue) String() string {
	return formatfloat(float64(p), 1) + "%"
}

func formatfloat(x float64, prec int) string {
	e := math.Pow10(prec)
	r := math.Round(x*e) / e
	return strconv.FormatFloat(r, 'f', -1, 64)
}

// Property is a benchmark configuration property.
type Property struct {
	Key   string
	Value Value
}

// String represents the property as a configuration line.
func (p Property) String() string {
	s := p.Key + ":"
	v := p.Value.String()
	if v != "" {
		s += " " + v
	}
	return s
}

// Validate the benchmark property complies with the Go Benchmark Data Format.
//
// Reference: https://github.com/golang/proposal/blob/d74d825331d9b16ee286ea77c0e4caeaf0efbe30/design/14313-benchmark-format.md#L101-L113
//
//	A configuration line is a key-value pair of the form
//
//		key: value
//
//	where key begins with a lower case character (as defined by `unicode.IsLower`),
//	contains no space characters (as defined by `unicode.IsSpace`)
//	nor upper case characters (as defined by `unicode.IsUpper`),
//	and one or more ASCII space or tab characters separate “key:” from “value.”
//	Conventionally, multiword keys are written with the words
//	separated by hyphens, as in cpu-speed.
//	There are no restrictions on value, except that it cannot contain a newline character.
//	Value can be omitted entirely, in which case the colon must still be
//	present, but need not be followed by a space.
//
func (p Property) Validate() error {
	if p.Key == "" {
		return errors.New("empty key")
	}

	for i, r := range p.Key {
		switch {
		case i == 0 && !unicode.IsLower(r):
			return errors.New("key starts with non lower case")
		case unicode.IsSpace(r):
			return errors.New("key contains space character")
		case unicode.IsUpper(r):
			return errors.New("key contains upper case character")
		}
	}

	if strings.ContainsRune(p.Value.String(), '\n') {
		return errors.New("value contains new line")
	}

	return nil
}

// Configuration is a set of benchmark properties.
type Configuration []Property

// Provider is a source of configuration properties.
type Provider interface {
	Configuration() (Configuration, error)
}

// ProviderFunc adapts a function to the Provider interface.
type ProviderFunc func() (Configuration, error)

// Configuration calls f.
func (f ProviderFunc) Configuration() (Configuration, error) {
	return f()
}

type Prefixed struct {
	prefix string
	c      Configuration
}

func NewPrefixed(prefix string) *Prefixed {
	return &Prefixed{
		prefix: prefix,
	}
}

func (p *Prefixed) Add(k string, v Value) {
	p.c = append(p.c, Property{Key: p.prefix + "-" + k, Value: v})
}

func (p *Prefixed) Configuration() (Configuration, error) {
	return p.c, nil
}
