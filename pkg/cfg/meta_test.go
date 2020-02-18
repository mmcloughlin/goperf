package cfg

import "testing"

func TestPropertyValidateOK(t *testing.T) {
	valid := []Property{
		{"k", "value"},
		{"split-words-with-hyphens", "valid"},
		{"empty-values-are-valid", ""},
	}
	for _, p := range valid {
		if err := p.Validate(); err != nil {
			t.Errorf("expected nil; got %q", err)
		}
	}
}

func TestPropertyValidateErrors(t *testing.T) {
	cases := []struct {
		Property     Property
		ErrorMessage string
	}{
		{
			Property:     Property{Key: "", Value: "value"},
			ErrorMessage: "empty key",
		},
		{
			Property:     Property{Key: "Key", Value: "value"},
			ErrorMessage: "key starts with non lower case",
		},
		{
			Property:     Property{Key: "cpu model", Value: "value"},
			ErrorMessage: "key contains space character",
		},
		{
			Property:     Property{Key: "cpuModel", Value: "value"},
			ErrorMessage: "key contains upper case character",
		},
		{
			Property:     Property{Key: "cpu-model", Value: "Brand: Intel\nFreq: 2.80GHz\n"},
			ErrorMessage: "value contains new line",
		},
	}
	for _, c := range cases {
		err := c.Property.Validate()
		if err == nil {
			t.Fatal("expected error; got nil")
		}
		if err.Error() != c.ErrorMessage {
			t.Errorf("got error %q; expect %q", err.Error(), c.ErrorMessage)
		}
	}
}
