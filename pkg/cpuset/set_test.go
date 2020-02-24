package cpuset

import (
	"strings"
	"testing"
)

func TestSetListFormatBidirectional(t *testing.T) {
	cases := []struct {
		List    string
		Members []uint
	}{
		{
			List:    "0-4,9",
			Members: []uint{0, 1, 2, 3, 4, 9},
		},
		{
			List:    "0-2,7,12-14",
			Members: []uint{0, 1, 2, 7, 12, 13, 14},
		},
		{
			List:    "10",
			Members: []uint{10},
		},
		{
			List:    "1-3",
			Members: []uint{1, 2, 3},
		},
		{
			List:    "",
			Members: []uint{},
		},
	}
	for _, c := range cases {
		set := NewSet(c.Members...)

		// Parse
		got, err := ParseList(c.List)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equals(set) {
			t.Errorf("ParseList(%v) = %v; expect %v", c.List, got, set)
		}

		// Format
		if s := set.FormatList(); s != c.List {
			t.Errorf("(%v).FormatList() = %v; expect %v", set, s, c.List)
		}
	}
}

func TestSetParseListErrors(t *testing.T) {
	cases := []struct {
		List           string
		ErrorSubstring string
	}{
		{"wat", `parse unsigned integer: "wat" invalid`},
		{"wat-2", `parse unsigned integer: "wat" invalid`},
		{"2-wat", `parse unsigned integer: "wat" invalid`},
		{"1,2-3-4,5", `range "2-3-4" has too many splits`},
		{"1,4-2,5", `start exceeds end in range "4-2"`},
	}
	for _, c := range cases {
		got, err := ParseList(c.List)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got != nil {
			t.Fatal("expected nil return with error")
		}
		if !strings.Contains(err.Error(), c.ErrorSubstring) {
			t.Errorf("expect error %q to contain %q", err, c.ErrorSubstring)
		}
	}
}
