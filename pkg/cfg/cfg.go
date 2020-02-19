package cfg

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/mmcloughlin/cb/internal/errutil"
)

// Validatable is something that can be validated.
type Validatable interface {
	Validate() error
}

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

func (p PercentageValue) Validate() error {
	if !(0 <= float64(p) && float64(p) <= 100) {
		return errors.New("percentage must be between 0 and 100")
	}
	return nil
}

func formatfloat(x float64, prec int) string {
	e := math.Pow10(prec)
	r := math.Round(x*e) / e
	return strconv.FormatFloat(r, 'f', -1, 64)
}

type Label struct {
	Key         string
	Description string
}

type Configuration []Entry

func (c Configuration) Validate() error {
	var errs errutil.Errors
	for _, e := range c {
		if err := e.Validate(); err != nil {
			errs.Add(err)
		}
	}
	return errs.Err()
}

type Entry struct {
	Label
	Value Value
	Sub   Configuration
}

func Property(k string, v Value) Entry {
	return Entry{
		Label: Label{Key: k},
		Value: v,
	}
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
func (e Entry) Validate() error {
	// Validate the key.
	if e.Key == "" {
		return errors.New("empty key")
	}

	for i, r := range e.Key {
		switch {
		case i == 0 && !unicode.IsLower(r):
			return errors.New("key starts with non lower case")
		case unicode.IsSpace(r):
			return errors.New("key contains space character")
		case unicode.IsUpper(r):
			return errors.New("key contains upper case character")
		case r == ':':
			return errors.New("key contains colon character")
		}
	}

	// Should have one of value or sub-config.
	hassub := len(e.Sub) > 0
	hasvalue := e.Value != nil
	switch {
	case hassub && hasvalue:
		return errors.New("entry has both sub-configuration and a value")
	case !hassub && !hasvalue:
		return errors.New("empty entry")
	}

	// Validate sub-config.
	if hassub {
		return e.Sub.Validate()
	}

	// Validate Value.
	if strings.ContainsRune(e.Value.String(), '\n') {
		return errors.New("value contains new line")
	}

	if v, ok := e.Value.(Validatable); ok {
		return v.Validate()
	}

	return nil
}

// Write configuration to the writer w.
func Write(w io.Writer, c Configuration) error {
	if err := c.Validate(); err != nil {
		return err
	}
	wr := &writer{Writer: w}
	wr.configuration(c)
	return wr.err
}

type writer struct {
	io.Writer
	prefix string
	err    error
}

func (w *writer) configuration(c Configuration) {
	for _, e := range c {
		w.entry(e)
	}
}

func (w *writer) entry(e Entry) {
	// Print value.
	if e.Value != nil {
		w.line(e.Key, e.Value.String())
		return
	}

	// Recurse into sub-config.
	save := w.prefix
	w.prefix = w.prefix + e.Key + "-"
	w.configuration(e.Sub)
	w.prefix = save
}

func (w *writer) line(k, v string) {
	k = w.prefix + k
	if v == "" {
		w.printf("%s:\n", k)
	} else {
		w.printf("%s: %s\n", k, v)
	}
}

func (w *writer) printf(format string, a ...interface{}) {
	if w.err != nil {
		return
	}
	_, w.err = fmt.Fprintf(w, format, a...)
}
