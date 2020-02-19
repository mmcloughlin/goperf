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
		KeyValue("k", StringValue("value")),
		KeyValue("split-words-with-hyphens", StringValue("valid")),
		KeyValue("empty-values-are-valid", StringValue("")),
	}
	for _, p := range valid {
		if err := p.Validate(); err != nil {
			t.Errorf("expected nil; got %q", err)
		}
	}
}

func TestEntryValidateErrors(t *testing.T) {
	cases := []struct {
		Entry        Entry
		ErrorMessage string
	}{
		{
			Entry:        KeyValue("", StringValue("value")),
			ErrorMessage: "empty key",
		},
		{
			Entry:        KeyValue("Key", StringValue("value")),
			ErrorMessage: "key starts with non lower case",
		},
		{
			Entry:        KeyValue("cpu model", StringValue("value")),
			ErrorMessage: "key contains space character",
		},
		{
			Entry:        KeyValue("cpuModel", StringValue("value")),
			ErrorMessage: "key contains upper case character",
		},
		{
			Entry:        KeyValue("cpu:model", StringValue("value")),
			ErrorMessage: "key contains colon character",
		},
		{
			Entry:        KeyValue("cpu-model", StringValue("Brand: Intel\nFreq: 2.80GHz\n")),
			ErrorMessage: "value contains new line",
		},
		{
			Entry:        KeyValue("used-percent", PercentageValue(120)),
			ErrorMessage: "percentage must be between 0 and 100",
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
