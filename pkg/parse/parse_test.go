package parse

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParse(t *testing.T) {
	lines := []string{
		"a: a1",
		"b: b1",
		"BenchmarkX/k1=v1/v2/k3=v3-8 100 42.42 unita 100 unitb",
		"b: b2",
		"c: c1",
		"BenchmarkY-4 100 37.6 unitc",
	}
	expect := &Collection{
		Results: []*Result{
			{
				FullName:   "BenchmarkX/k1=v1/v2/k3=v3-8",
				Name:       "X",
				Parameters: map[string]string{"k1": "v1", "sub2": "v2", "k3": "v3", "gomaxprocs": "8"},
				Labels:     map[string]string{"a": "a1", "b": "b1"},
				Iterations: 100,
				Value:      42.42,
				Unit:       "unita",
				Line:       3,
			},
			{
				FullName:   "BenchmarkX/k1=v1/v2/k3=v3-8",
				Name:       "X",
				Parameters: map[string]string{"k1": "v1", "sub2": "v2", "k3": "v3", "gomaxprocs": "8"},
				Labels:     map[string]string{"a": "a1", "b": "b1"},
				Iterations: 100,
				Value:      100,
				Unit:       "unitb",
				Line:       3,
			},
			{
				FullName:   "BenchmarkY-4",
				Name:       "Y",
				Parameters: map[string]string{"gomaxprocs": "4"},
				Labels:     map[string]string{"a": "a1", "b": "b2", "c": "c1"},
				Iterations: 100,
				Value:      37.6,
				Unit:       "unitc",
				Line:       6,
			},
		},
	}

	// Prepare input.
	buf := bytes.NewBuffer(nil)
	for _, line := range lines {
		fmt.Fprintln(buf, line)
	}

	// Parse.
	got, err := Reader(buf)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Logf("diff =\n%s", diff)
		t.FailNow()
	}
}

func TestParseTestdata(t *testing.T) {
	filenames, err := filepath.Glob("testdata/*.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, filename := range filenames {
		filename := filename // scopelint
		t.Run(filepath.Base(filename), func(t *testing.T) {
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}

			c, err := Bytes(b)
			if err != nil {
				t.Fatal(err)
			}

			for _, e := range c.Errors {
				t.Log(e)
			}

			if len(c.Results) == 0 {
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
