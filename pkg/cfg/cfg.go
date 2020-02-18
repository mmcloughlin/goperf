package cfg

import (
	"errors"
	"strings"
	"unicode"
)

// Property is a benchmark configuration property.
type Property struct {
	Key   string
	Value string
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

	if strings.ContainsRune(p.Value, '\n') {
		return errors.New("value contains new line")
	}

	return nil
}
