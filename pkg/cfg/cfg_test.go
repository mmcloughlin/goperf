package cfg

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func TestEntryTypes(t *testing.T) {
	var (
		_ Entry = PropertyEntry{}
		_ Entry = SectionEntry{}
		_ Entry = Property("key", "doc", StringValue("value"))
		_ Entry = KeyValue("key", StringValue("value"))
	)
}

func TestSectionEntryIsProvider(t *testing.T) {
	var _ Provider = SectionEntry{}
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
			ErrorMessage: "empty",
		},
		{
			Entry:        KeyValue("Key", StringValue("value")),
			ErrorMessage: "starts with non lower case",
		},
		{
			Entry:        KeyValue("cpu model", StringValue("value")),
			ErrorMessage: "contains space character",
		},
		{
			Entry:        KeyValue("cpuModel", StringValue("value")),
			ErrorMessage: "contains upper case character",
		},
		{
			Entry:        KeyValue("cpu:model", StringValue("value")),
			ErrorMessage: "contains colon character",
		},
		{
			Entry:        KeyValue("cpu-model", StringValue("Brand: Intel\nFreq: 2.80GHz\n")),
			ErrorMessage: "value contains new line",
		},
		{
			Entry:        KeyValue("used-percent", PercentageValue(120)),
			ErrorMessage: "percentage must be between 0 and 100",
		},
		{
			Entry:        KeyValue("procstat-policy", StringValue("SCHED_RR"), "PerfCritical"),
			ErrorMessage: `tag "PerfCritical": starts with non lower case`,
		},
		{
			Entry:        KeyValue("procstat-policy", StringValue("SCHED_RR"), "left[bracket"),
			ErrorMessage: `tag "left[bracket": contains left square bracket character`,
		},
		{
			Entry:        KeyValue("procstat-policy", StringValue("SCHED_RR"), "right]bracket"),
			ErrorMessage: `tag "right]bracket": contains right square bracket character`,
		},
		{
			Entry:        KeyValue("procstat-policy", StringValue("SCHED_RR"), "comma,separated"),
			ErrorMessage: `tag "comma,separated": contains comma character`,
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

func TestWrite(t *testing.T) {
	cases := []struct {
		Configuration Configuration
		Expect        string
	}{
		{
			Configuration: Configuration{KeyValue("key", StringValue(""))},
			Expect:        "key:\n",
		},
		{
			Configuration: Configuration{KeyValue("key", StringValue("value"))},
			Expect:        "key: value\n",
		},
		{
			Configuration: Configuration{
				KeyValue("key", StringValue("value"), "tag"),
			},
			Expect: "key: value [tag]\n",
		},
		{
			Configuration: Configuration{
				KeyValue("key", StringValue("value"), "tag1", "tag2", "tag3"),
			},
			Expect: "key: value [tag1,tag2,tag3]\n",
		},
	}
	for _, c := range cases {
		buf := bytes.NewBuffer(nil)
		if err := Write(buf, c.Configuration); err != nil {
			t.Fatal(err)
		}
		got := buf.String()
		if diff := cmp.Diff(c.Expect, got); diff != "" {
			t.Logf("expect\n%s", c.Expect)
			t.Logf("got\n%s", got)
			t.Logf("diff\n%s", diff)
			t.Fail()
		}
	}
}
