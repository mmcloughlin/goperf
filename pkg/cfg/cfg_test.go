package cfg

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/cb/pkg/parse"
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
		KeyValue("multiplewords", StringValue("valid")),
		KeyValue("emptyvaluesarevalid", StringValue("")),
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
			Entry:        KeyValue("section-separator", StringValue("value")),
			ErrorMessage: `contains section separator '-'`,
		},
		{
			Entry:        KeyValue("cpumodel", StringValue("Brand: Intel\nFreq: 2.80GHz\n")),
			ErrorMessage: "value contains new line",
		},
		{
			Entry:        KeyValue("usedpercent", PercentageValue(120)),
			ErrorMessage: "percentage must be between 0 and 100",
		},
		{
			Entry:        KeyValue("policy", StringValue("SCHED_RR"), "PerfCritical"),
			ErrorMessage: `tag "PerfCritical": starts with non lower case`,
		},
		{
			Entry:        KeyValue("policy", StringValue("SCHED_RR"), "left[bracket"),
			ErrorMessage: `tag "left[bracket": contains left square bracket character`,
		},
		{
			Entry:        KeyValue("policy", StringValue("SCHED_RR"), "right]bracket"),
			ErrorMessage: `tag "right]bracket": contains right square bracket character`,
		},
		{
			Entry:        KeyValue("policy", StringValue("SCHED_RR"), "comma,separated"),
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
		{
			Configuration: Configuration{
				Section(
					"section",
					"some configuration section",
					KeyValue("a", IntValue(1)),
					KeyValue("b", IntValue(2)),
					KeyValue("c", IntValue(3)),
				),
			},
			Expect: strings.Join([]string{
				"section-a: 1",
				"section-b: 2",
				"section-c: 3",
				"",
			}, "\n"),
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

func TestParseValueTags(t *testing.T) {
	cases := []struct {
		Input string
		Value string
		Tags  []Tag
	}{
		{
			Input: "hello world",
			Value: "hello world",
			Tags:  nil,
		},
		{
			Input: "hello world [perf]",
			Value: "hello world",
			Tags:  []Tag{TagPerfCritical},
		},
		{
			Input: "hello world [a,b,c]",
			Value: "hello world",
			Tags:  []Tag{"a", "b", "c"},
		},
		{
			Input: "hello world\t  \t  [a,b,c]",
			Value: "hello world",
			Tags:  []Tag{"a", "b", "c"},
		},
		{
			Input: " [a]",
			Value: "",
			Tags:  []Tag{"a"},
		},
		{
			Input: "[a]",
			Value: "",
			Tags:  []Tag{"a"},
		},
		{
			Input: "hello world [these are not tags]",
			Value: "hello world [these are not tags]",
			Tags:  nil,
		},
		{
			Input: "hello world [this,has,an,invalid,,empty,tag]",
			Value: "hello world [this,has,an,invalid,,empty,tag]",
			Tags:  nil,
		},
	}
	for _, c := range cases {
		v, tags := ParseValueTags(c.Input)
		if v != c.Value || !reflect.DeepEqual(tags, c.Tags) {
			t.Errorf("ParseValueTags(%s) = %v, %v; expect %v, %v", c.Input, v, tags, c.Value, c.Tags)
		}
	}
}

func TestWriteParseTagsRoundtrip(t *testing.T) {
	// Write configuration lines.
	expect := []Tag{"a", "b", "c"}
	c := Configuration{
		KeyValue("key", StringValue("value"), expect...),
	}
	buf := bytes.NewBuffer(nil)
	err := Write(buf, c)
	if err != nil {
		t.Fatal(err)
	}

	// Write a fake benchmark line, so we can use the parser.
	fmt.Fprintln(buf, "BenchmarkEncodeDigitsSpeed1e4-8   	      30	    482808 ns/op")

	// Parse out.
	collection, err := parse.Reader(buf)
	if err != nil {
		t.Fatal(err)
	}

	// Extract label.
	if len(collection.Results) != 1 {
		t.Fatal("expected one result")
	}
	s := collection.Results[0].Labels["key"]

	// Parse out value and tags.
	v, tags := ParseValueTags(s)
	if v != "value" {
		t.Errorf(`expected "value" got %q`, v)
	}

	if !reflect.DeepEqual(expect, tags) {
		t.Errorf(`expected tags %v got %v`, expect, tags)
	}
}
