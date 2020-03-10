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

// StringsValue is a list of strings.
type StringsValue []string

func (s StringsValue) String() string { return strings.Join(s, ",") }

// Float64Value is a double-precision floating point.
type Float64Value float64

func (x Float64Value) String() string { return strconv.FormatFloat(float64(x), 'f', -1, 64) }

// IntValue is an integer.
type IntValue int

func (x IntValue) String() string { return strconv.Itoa(int(x)) }

// BoolValue is a boolean.
type BoolValue bool

func (b BoolValue) String() string { return strconv.FormatBool(bool(b)) }

// BytesValue represents bytes.
type BytesValue uint64

func (b BytesValue) String() string {
	return formatunit(float64(b), 1024, []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}, 2)
}

// BytesValue represents bytes.
type FrequencyValue float64

func (f FrequencyValue) String() string {
	return formatunit(float64(f), 1000, []string{"Hz", "KHz", "MHz", "GHz"}, 2)
}

func formatunit(x, mult float64, units []string, prec int) string {
	i := 0
	for x >= mult && i+1 < len(units) {
		x /= mult
		i++
	}
	return formatfloat(x, prec) + " " + units[i]
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

// Label for a configuration property or section.
type Labeled interface {
	Key() Key
	Doc() string
}

// Entry is the base type for configuration entries. (Note this is a sealed
// interface, it may not be implemented outside this package.)
type Entry interface {
	Labeled
	Validatable

	entry() // sealed
}

type label struct {
	key Key
	doc string
}

func Label(k Key, doc string) Labeled {
	return label{key: k, doc: doc}
}

func (l label) Key() Key    { return l.key }
func (l label) Doc() string { return l.doc }

// Tag for a configuration property.
type Tag string

// Standard tag defintions.
const (
	TagNone         Tag = ""
	TagPerfCritical Tag = "perf"
)

type PropertyEntry struct {
	Labeled
	Value Value
	Tags  []Tag
}

func (PropertyEntry) entry() {}

func Property(k Key, doc string, v Value, tags ...Tag) PropertyEntry {
	return PropertyEntry{
		Labeled: Label(k, doc),
		Value:   v,
		Tags:    tags,
	}
}

// PerfProperty builds a property tagged as performance critical.
func PerfProperty(k Key, doc string, v Value, tags ...Tag) PropertyEntry {
	return Property(k, doc, v, append([]Tag{TagPerfCritical}, tags...)...)
}

// KeyValue builds an undocumented property.
func KeyValue(k Key, v Value, tags ...Tag) PropertyEntry {
	return Property(k, "", v, tags...)
}

// SectionEntry is a nested configuration.
type SectionEntry struct {
	Labeled
	Sub Configuration
}

func (SectionEntry) entry() {}

func Section(k Key, doc string, entries ...Entry) SectionEntry {
	return SectionEntry{
		Labeled: Label(k, doc),
		Sub:     Configuration(entries),
	}
}

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
		return errors.New("empty")
	}

	for i, r := range k {
		switch {
		case i == 0 && !unicode.IsLower(r):
			return errors.New("starts with non lower case")
		case unicode.IsSpace(r):
			return errors.New("contains space character")
		case unicode.IsUpper(r):
			return errors.New("contains upper case character")
		case r == ':':
			return errors.New("contains colon character")
		}
	}

	return nil
}

// Validate that tag has the allowed form. Must meet the same requirements as a
// key, and in addition may not contain square brackets or commas.
func (t Tag) Validate() error {
	// Validate as a key.
	if err := Key(t).Validate(); err != nil {
		return err
	}

	// Check additional disallowed characters.
	disallowed := map[rune]string{
		'[': "left square bracket",
		']': "right square bracket",
		',': "comma",
	}
	for _, r := range t {
		if _, ok := disallowed[r]; ok {
			return fmt.Errorf("contains %s character", disallowed[r])
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
func (p PropertyEntry) Validate() error {
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

	// Validate tags.
	for _, t := range p.Tags {
		if err := t.Validate(); err != nil {
			return fmt.Errorf("tag %q: %w", t, err)
		}
	}

	return nil
}

// Validate confirms the section key and sub-configuration are valid.
func (s SectionEntry) Validate() error {
	// Validate key.
	if err := s.Key().Validate(); err != nil {
		return err
	}

	// Validate sub-config.
	return s.Sub.Validate()
}

// Generator generates Configuration.
type Generator interface {
	Configuration() (Configuration, error)
}

// GeneratorFunc adapts a function to the Generator interface.
type GeneratorFunc func() (Configuration, error)

// Configuration calls g.
func (g GeneratorFunc) Configuration() (Configuration, error) {
	return g()
}

// Provider is a source of configuration.
type Provider interface {
	Labeled
	Generator
	Available() bool
}

// Available returns true (satisfies the Provider interface).
func (s SectionEntry) Available() bool { return true }

// Configuration satisfies the Provider interface.
func (s SectionEntry) Configuration() (Configuration, error) { return s.Sub, nil }

type provider struct {
	Labeled
	Generator
}

// NewProvider builds a Provider from a Generator.
func NewProvider(k Key, doc string, g Generator) Provider {
	return provider{
		Labeled:   Label(k, doc),
		Generator: g,
	}
}

// NewProviderFunc builds a Provider from a function.
func NewProviderFunc(k Key, doc string, f func() (Configuration, error)) Provider {
	return NewProvider(k, doc, GeneratorFunc(f))
}

func (p provider) Available() bool { return true }

// Providers is a list of providers.
type Providers []Provider

// Available returns true. Note that sub-providers will be checked for
// availability when Configuration() is called.
func (p Providers) Available() bool { return true }

// Keys returns all keys in the provider list.
func (p Providers) Keys() []string {
	keys := make([]string, len(p))
	for i := range p {
		keys[i] = string(p[i].Key())
	}
	return keys
}

// FilterAvailable returns the available sub-providers.
func (p Providers) FilterAvailable() Providers {
	a := make(Providers, 0, len(p))
	for i := range p {
		if p[i].Available() {
			a = append(a, p[i])
		}
	}
	return a
}

// Select returns the subset of providers with the given keys.
func (p Providers) Select(keys ...string) (Providers, error) {
	m := map[string]Provider{}
	for i := range p {
		m[string(p[i].Key())] = p[i]
	}

	s := make(Providers, len(keys))
	for i, k := range keys {
		if _, ok := m[k]; !ok {
			return nil, fmt.Errorf("provider %q not found", k)
		}
		s[i] = m[k]
	}
	return s, nil
}

// Configuration gathers configuration from all providers.
func (p Providers) Configuration() (Configuration, error) {
	c := make(Configuration, len(p))
	for i := range p {
		e, err := providerentry(p[i])
		if err != nil {
			return nil, err
		}
		c[i] = e
	}
	return c, nil
}

func providerentry(p Provider) (Entry, error) {
	section := SectionEntry{
		Labeled: p,
	}

	if !p.Available() {
		section.Sub = Configuration{
			Property(
				"available",
				fmt.Sprintf("availability of the %s provider", p.Key()),
				BoolValue(false),
			),
		}
	} else {
		sub, err := p.Configuration()
		if err != nil {
			return nil, err
		}
		section.Sub = sub
	}

	return section, nil
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
	case PropertyEntry:
		w.property(en)
	case SectionEntry:
		save := w.prefix
		w.prefix = w.prefix + en.Key() + "-"
		w.configuration(en.Sub)
		w.prefix = save
	default:
		w.seterr(errutil.UnexpectedType(e))
	}
}

func (w *writer) property(p PropertyEntry) {
	w.printf("%s:", w.prefix+p.Key())

	if v := p.Value.String(); v != "" {
		w.printf(" %s", v)
	}

	if len(p.Tags) > 0 {
		sep := " ["
		for _, t := range p.Tags {
			w.printf("%s%s", sep, t)
			sep = ","
		}
		w.printf("]")
	}

	w.printf("\n")
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
