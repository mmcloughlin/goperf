package flags

import (
	"fmt"
	"strings"
)

// Strings is a list of strings satisfying the flag.Value interface.
type Strings []string

func (a Strings) String() string {
	return strings.Join(a, ",")
}

// Set splits the comma-separated string s.
func (a *Strings) Set(s string) error {
	*a = strings.Split(s, ",")
	return nil
}

// Param is a key-value pair.
type Param struct {
	Key   string
	Value string
}

// Params is a collection of key value pairs. Satisfies the flag.Value interface.
type Params []Param

// ParseParams parses a string into a key-value pair map. Expects a comma-separated list of key-value pairs, for example:
//
//	a=1,b=2,c=3
func ParseParams(s string) (Params, error) {
	params := Params{}
	for _, kv := range strings.Split(s, ",") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("parameter %q is not a key-value pair", kv)
		}
		params = append(params, Param{
			Key:   parts[0],
			Value: parts[1],
		})
	}
	return params, nil
}

func (p Params) String() string {
	kvs := make([]string, len(p))
	for i := range p {
		kvs[i] = p[i].Key + "=" + p[i].Value
	}
	return strings.Join(kvs, ",")
}

func (p *Params) Set(s string) error {
	ps, err := ParseParams(s)
	if err != nil {
		return err
	}
	*p = ps
	return nil
}

func (p Params) Map() map[string]string {
	m := map[string]string{}
	for i := range p {
		m[p[i].Key] = p[i].Value
	}
	return m
}

// TypeParams is a concise configuration of some object, with a type and
// parameters. Satisfies the flag.Value interface.
type TypeParams struct {
	Type   string
	Params Params
}

// ParseTypeParams parses a string into type and parameters. For example:
//
//	type:a=1,b=2,c=3
func ParseTypeParams(s string) (*TypeParams, error) {
	parts := strings.SplitN(s, ":", 2)

	if len(parts) == 1 {
		return &TypeParams{
			Type: parts[0],
		}, nil
	}

	params, err := ParseParams(parts[1])
	if err != nil {
		return nil, err
	}

	return &TypeParams{
		Type:   parts[0],
		Params: params,
	}, nil
}

func (t *TypeParams) String() string {
	s := t.Type
	if len(t.Params) > 0 {
		s += ":" + t.Params.String()
	}
	return s
}

func (t *TypeParams) Set(s string) error {
	tp, err := ParseTypeParams(s)
	if err != nil {
		return err
	}
	*t = *tp
	return nil
}
