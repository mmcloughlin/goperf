package cli

import (
	"reflect"
	"testing"
)

func TestParseTypeParams(t *testing.T) {
	cases := []struct {
		Input  string
		Expect *TypeParams
	}{
		{
			Input:  "",
			Expect: &TypeParams{Type: "", Params: map[string]string{}},
		},
		{
			Input:  "type",
			Expect: &TypeParams{Type: "type", Params: map[string]string{}},
		},
		{
			Input:  "type:a=1",
			Expect: &TypeParams{Type: "type", Params: map[string]string{"a": "1"}},
		},
		{
			Input:  "type:a=1,b=2",
			Expect: &TypeParams{Type: "type", Params: map[string]string{"a": "1", "b": "2"}},
		},
		{
			Input:  "type:a=1,b=2,c=d=3,e=4",
			Expect: &TypeParams{Type: "type", Params: map[string]string{"a": "1", "b": "2", "c": "d=3", "e": "4"}},
		},
	}
	for _, c := range cases {
		got, err := ParseTypeParams(c.Input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.Expect, got) {
			t.Errorf("ParseTypeParams(%q) = %#v; expect %#v", c.Input, got, c.Expect)
		}
	}
}

func TestParseTypeParamsErrors(t *testing.T) {
	cases := []struct {
		Input        string
		ErrorMessage string
	}{
		{
			Input:        "type:abc",
			ErrorMessage: "parameter \"abc\" is not a key-value pair",
		},
	}
	for _, c := range cases {
		got, err := ParseTypeParams(c.Input)
		if err == nil {
			t.Fatal("expected error; got nil")
		}
		if got != nil {
			t.Fatalf("expected nil response with error")
		}
		if err.Error() != c.ErrorMessage {
			t.Errorf("got error %q; expected %q", err.Error(), c.ErrorMessage)
		}
	}
}
