package parse

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseTestdata(t *testing.T) {
	filenames, err := filepath.Glob("testdata/*.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, filename := range filenames {
		t.Run(filepath.Base(filename), func(t *testing.T) {
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}

			rs, err := Bytes(b)
			if err != nil {
				t.Fatal(err)
			}

			if len(rs) == 0 {
				t.Fatal("no results parsed")
			}
		})
	}
}

func TestParseResultLine(t *testing.T) {
	line := "BenchmarkDecodeTwainCompress 100 138516 ns/op 72.19 MB/s 40849 B/op 15 allocs/op"
	got, err := parseresultline(line)
	if err != nil {
		t.Fatal(err)
	}

	expect := &resultline{
		name:       "BenchmarkDecodeTwainCompress",
		iterations: 100,
		measurements: []measurement{
			{138516, "ns/op"},
			{72.19, "MB/s"},
			{40849, "B/op"},
			{15, "allocs/op"},
		},
	}

	if !reflect.DeepEqual(expect, got) {
		t.FailNow()
	}
}

func TestParseResultLineErrors(t *testing.T) {
	cases := []string{
		"BenchmarkHello 42", // not enough fields
		"BenchmarkDecodeTwainCompress 138516 ns/op 72.19 MB/s 40849 B/op 15 allocs/op",            // odd number of fields
		"SomethingDecodeTwainCompress 100 138516 ns/op 72.19 MB/s 40849 B/op 15 allocs/op",        // wrong name prefix
		"BenchmarkDecodeTwainCompress iterations 138516 ns/op 72.19 MB/s 40849 B/op 15 allocs/op", // bad iterations field
		"BenchmarkDecodeTwainCompress 100 138516 ns/op 72.19 MB/s float B/op 15 allocs/op",        // bad value field
	}
	for _, c := range cases {
		r, err := parseresultline(c)
		t.Logf("input = %q", c)
		if r != nil || err == nil {
			t.FailNow()
		}
		t.Logf("error = %q", err.Error())
	}
}
