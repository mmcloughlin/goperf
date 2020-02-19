package cfg

import "testing"

func TestBytesValue(t *testing.T) {
	cases := []struct {
		Bytes  BytesValue
		Expect string
	}{
		{0, "0 B"},
		{42, "42 B"},
		{43150, "42.14 KiB"},
		{44185130, "42.14 MiB"},
		{45245572883, "42.14 GiB"},
		{46331466632323, "42.14 TiB"},
		{47287796087390208, "42 PiB"},
		{^BytesValue(0), "16 EiB"},

		// Float precision.
		{2048, "2 KiB"},     // 0
		{73114, "71.4 KiB"}, // 1
		{3217, "3.14 KiB"},  // 2
	}
	for _, c := range cases {
		if got := c.Bytes.String(); got != c.Expect {
			t.Errorf("(%d).String() = %q; expect %q", c.Bytes, got, c.Expect)
		}
	}
}

func TestEntryValidateOK(t *testing.T) {
	valid := []Entry{
		Property("k", StringValue("value")),
		Property("split-words-with-hyphens", StringValue("valid")),
		Property("empty-values-are-valid", StringValue("")),
	}
	for _, p := range valid {
		if err := p.Validate(); err != nil {
			t.Errorf("expected nil; got %q", err)
		}
	}
}

func TestPropertyValidateErrors(t *testing.T) {
	cases := []struct {
		Entry        Entry
		ErrorMessage string
	}{
		{
			Entry:        Property("", StringValue("value")),
			ErrorMessage: "empty key",
		},
		{
			Entry:        Property("Key", StringValue("value")),
			ErrorMessage: "key starts with non lower case",
		},
		{
			Entry:        Property("cpu model", StringValue("value")),
			ErrorMessage: "key contains space character",
		},
		{
			Entry:        Property("cpuModel", StringValue("value")),
			ErrorMessage: "key contains upper case character",
		},
		{
			Entry:        Property("cpu:model", StringValue("value")),
			ErrorMessage: "key contains colon character",
		},
		{
			Entry:        Property("cpu-model", StringValue("Brand: Intel\nFreq: 2.80GHz\n")),
			ErrorMessage: "value contains new line",
		},
		{
			Entry:        Property("used-percent", PercentageValue(120)),
			ErrorMessage: "percentage must be between 0 and 100",
		},
		{
			Entry: Entry{
				Label: Label{Key: "empty"},
				Value: nil,
				Sub:   nil,
			},
			ErrorMessage: "empty entry",
		},
		{
			Entry: Entry{
				Label: Label{Key: "both"},
				Value: StringValue("value"),
				Sub: Configuration{
					Property("a", StringValue("1")),
					Property("b", StringValue("2")),
					Property("c", StringValue("3")),
				},
			},
			ErrorMessage: "entry has both sub-configuration and a value",
		},
	}
	for _, c := range cases {
		err := c.Entry.Validate()
		if err == nil {
			t.Fatal("expected error; got nil")
		}
		if err.Error() != c.ErrorMessage {
			t.Errorf("got error %q; expect %q", err.Error(), c.ErrorMessage)
		}
	}
}
