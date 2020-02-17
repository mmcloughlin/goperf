package cli

import (
	"fmt"
	"strings"
)

// TypeParams is a concise configuration of some object, with a type and parameters.
type TypeParams struct {
	Type   string
	Params map[string]string
}

// ParseTypeParams parses a string into type and parameters. For example:
//
//	type:a=1,b=2,c=3
func ParseTypeParams(s string) (*TypeParams, error) {
	parts := strings.SplitN(s, ":", 2)

	if len(parts) == 1 {
		return &TypeParams{
			Type:   parts[0],
			Params: map[string]string{},
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

// ParseParams parses a string into a key-value pair map. Expects a comma-separated list of key-value pairs, for example:
//
//	a=1,b=2,c=3
func ParseParams(s string) (map[string]string, error) {
	params := map[string]string{}
	for _, kv := range strings.Split(s, ",") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("parameter %q is not a key-value pair", kv)
		}
		params[parts[0]] = parts[1]
	}
	return params, nil
}
