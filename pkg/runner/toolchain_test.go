package runner

import (
	"reflect"
	"runtime"
	"testing"
)

func TestNewToolchain(t *testing.T) {
	cases := []struct {
		Type   string
		Params map[string]string
		Expect Toolchain
	}{
		{
			Type: "snapshot",
			Params: map[string]string{
				"builder_type": "linux-amd64",
				"revision":     "3eab754cd061bf90ee7b540546bc0863f3ad1d85",
			},
			Expect: NewSnapshot("linux-amd64", "3eab754cd061bf90ee7b540546bc0863f3ad1d85"),
		},
		{
			Type: "release",
			Params: map[string]string{
				"version": "go1.13.8",
				"arch":    "amd64",
			},
			Expect: NewRelease("go1.13.8", runtime.GOOS, "amd64"),
		},
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.Type, func(t *testing.T) {
			tc, err := NewToolchain(c.Type, c.Params)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(c.Expect, tc) {
				t.Fatal("toolchain mismatch")
			}
		})
	}
}

func TestNewToolchainErrors(t *testing.T) {
	cases := []struct {
		Name         string
		Type         string
		Params       map[string]string
		ErrorMessage string
	}{
		{
			Name:         "unknown_type",
			Type:         "idk",
			Params:       map[string]string{},
			ErrorMessage: "unknown toolchain type: \"idk\"",
		},
		{
			Name:         "single_missing_field",
			Type:         "snapshot",
			Params:       map[string]string{"builder_type": "linux-amd64"},
			ErrorMessage: "missing parameter: revision",
		},
		{
			Name:         "multiple_missing_fields",
			Type:         "snapshot",
			Params:       map[string]string{},
			ErrorMessage: "missing parameters: builder_type, revision",
		},
		{
			Name: "single_extra_field",
			Type: "snapshot",
			Params: map[string]string{
				"builder_type": "linux-amd64",
				"revision":     "3eab754cd061bf90ee7b540546bc0863f3ad1d85",
				"idk":          "wat",
			},
			ErrorMessage: "unknown parameter: idk",
		},
		{
			Name: "multiple_extra_fields",
			Type: "snapshot",
			Params: map[string]string{
				"builder_type": "linux-amd64",
				"revision":     "3eab754cd061bf90ee7b540546bc0863f3ad1d85",
				"idk":          "wat",
				"jk":           "lol",
			},
			ErrorMessage: "unknown parameters: idk, jk",
		},
	}
	for _, c := range cases {
		c := c // scopelint
		t.Run(c.Name, func(t *testing.T) {
			tc, err := NewToolchain(c.Type, c.Params)
			if tc != nil {
				t.Fatal("expected nil toolchain")
			}
			if err == nil {
				t.Fatal("expected error; got nil")
			}
			if err.Error() != c.ErrorMessage {
				t.Fatalf("got error %q; expect %q", err.Error(), c.ErrorMessage)
			}
		})
	}
}
