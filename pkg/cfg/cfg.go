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

// StringValue is a string constant.
type StringValue string

func (s StringValue) String() string { return string(s) }

// BytesValue represents bytes.
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

// PercentageValue represents a percentage, therefore must be in the range 0 to 100.
type PercentageValue float64

func (p PercentageValue) String() string {
	return formatfloat(float64(p), 1) + "%"
}

// Validate checks the p is between 0 and 100.
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

// Configuration is a nested key-value structure. It is a list of entries, where
// each entry is either a key-value property or a section containing a nested
// config.
type Configuration []Entry

// Key is an identifier for a config property or section.
type Key string

// Entry is the base type for configuration entries. (Note this is a sealed
// interface, it may not be implemented outside this package.)
type Entry interface {
	Validatable

	Key() Key
	Documentation() string

	entry() // sealed
}

// Label for a configuration property or section.
type Label struct {
	Name Key
	Doc  string
}

// Key returns the label name.
func (l Label) Key() Key { return l.Name }

// Documentation returns human-readable description of the labeled item.
func (l Label) Documentation() string { return l.Doc }

// Property is a key-value pair.
type Property struct {
	Label
	Value Value
}

// KeyValue builds an undocumented Property.
func KeyValue(k Key, v Value) Property {
	return Property{
		Label: Label{Name: k},
		Value: v,
	}
}

// Section is a nested configuration.
type Section struct {
	Label
	Sub Configuration
}

func (Property) entry() {}
func (Section) entry()  {}

// Validate checks that all entries are valid.
func (c Configuration) Validate() error {
	var errs errutil.Errors
	for _, e := range c {
		if err := e.Validate(); err != nil {
			errs.Add(err)
		}
	}
	return errs.Err()
}

// Validate that the key conforms to the Go Benchmark Data Format.
//
// Reference: https://github.com/golang/proposal/blob/d74d825331d9b16ee286ea77c0e4caeaf0efbe30/design/14313-benchmark-format.md#L101-L110
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
//
func (k Key) Validate() error {
	if k == "" {
		return errors.New("empty key")
	}

	for i, r := range k {
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

	return nil
}

// Validate the property conforms to the Go Benchmark Data Format. This checks the key as well as the value, as described below.
//
// Reference: https://github.com/golang/proposal/blob/d74d825331d9b16ee286ea77c0e4caeaf0efbe30/design/14313-benchmark-format.md#L111-L113
//
//	There are no restrictions on value, except that it cannot contain a newline character.
//	Value can be omitted entirely, in which case the colon must still be
//	present, but need not be followed by a space.
//
// In addition, if the property value is Validatable, its Validate method will be called.
func (p Property) Validate() error {
	// Validate key.
	if err := p.Key().Validate(); err != nil {
		return err
	}

	// Validate Value.
	if strings.ContainsRune(p.Value.String(), '\n') {
		return errors.New("value contains new line")
	}

	if v, ok := p.Value.(Validatable); ok {
		return v.Validate()
	}

	return nil
}

// Validate confirms the section key and sub-configuration are valid.
func (s Section) Validate() error {
	// Validate key.
	if err := s.Key().Validate(); err != nil {
		return err
	}

	// Validate sub-config.
	return s.Sub.Validate()
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
	prefix Key
	err    error
}

func (w *writer) configuration(c Configuration) {
	for _, e := range c {
		w.entry(e)
	}
}

func (w *writer) entry(e Entry) {
	switch en := e.(type) {
	case Property:
		w.line(en.Key(), en.Value.String())
	case Section:
		save := w.prefix
		w.prefix = w.prefix + en.Key() + "-"
		w.configuration(en.Sub)
		w.prefix = save
	default:
		w.seterr(errutil.UnexpectedType(e))
	}
}

func (w *writer) line(k Key, v string) {
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
	_, err := fmt.Fprintf(w, format, a...)
	w.seterr(err)
}

func (w *writer) seterr(err error) {
	if w.err == nil {
		w.err = err
	}
}
