package cpuset

import (
	"testing"
)

func TestSetEqualsEqual(t *testing.T) {
	a := NewSet(1, 2, 4, 8)
	b := NewSet(1, 2, 4, 8)
	if !a.Equals(b) {
		t.Fail()
	}
}

func TestSetEqualsDifferentSize(t *testing.T) {
	a := NewSet(1, 2, 4)
	b := NewSet(1, 2, 4, 8)
	if a.Equals(b) {
		t.Fail()
	}
}

func TestSetEqualsDifferentElements(t *testing.T) {
	a := NewSet(1, 2, 4, 8)
	b := NewSet(1, 2, 5, 8)
	if a.Equals(b) {
		t.Fail()
	}
}

func TestSetContainsTrue(t *testing.T) {
	a := NewSet(1, 4, 8)
	b := NewSet(1, 2, 4, 8)
	if !b.Contains(a) {
		t.Fail()
	}
}

func TestSetContainsLarger(t *testing.T) {
	a := NewSet(1, 2, 4, 8, 16)
	b := NewSet(1, 2, 4, 8)
	if b.Contains(a) {
		t.Fail()
	}
}

func TestSetContainsFalse(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(1, 2, 4, 8)
	if b.Contains(a) {
		t.Fail()
	}
}

func TestSetMaskFormatBidirectional(t *testing.T) {
	cases := []struct {
		Mask    string
		Members []uint
	}{
		{
			Mask:    "00000000",
			Members: []uint{},
		},
		{
			Mask:    "00000001",
			Members: []uint{0},
		},
		{
			Mask:    "40000000,00000000,00000000",
			Members: []uint{94},
		},
		{
			Mask:    "00000001,00000000,00000000",
			Members: []uint{64},
		},
		{
			Mask:    "000000ff,00000000",
			Members: []uint{32, 33, 34, 35, 36, 37, 38, 39},
		},
		{
			Mask:    "000e3862",
			Members: []uint{1, 5, 6, 11, 12, 13, 17, 18, 19},
		},
		{
			Mask:    "00000001,00000001,00010117",
			Members: []uint{0, 1, 2, 4, 8, 16, 32, 64},
		},
	}
	for _, c := range cases {
		set := NewSet(c.Members...)

		// Parse
		got, err := ParseMask(c.Mask)
		if err != nil {
			t.Fatal(err)
		}
		if !got.Equals(set) {
			t.Errorf("ParseMask(%v) = %v; expect %v", c.Mask, got, set)
		}

		// Format
		if s := set.FormatMask(); s != c.Mask {
			t.Errorf("(%v).FormatMask() = %v; expect %v", set, s, c.Mask)
		}
	}
}

func TestSetParseMaskErrors(t *testing.T) {
	cases := []struct {
		Mask  string
		Error string
	}{
		{"0", `parsing mask word "0": expected 8 hex characters`},
		{"00000100,qwertyui,11223344", `parsing mask word "qwertyui": invalid`},
		{"0x112233", `parsing mask word "0x112233": invalid`},
	}
	for _, c := range cases {
		got, err := ParseMask(c.Mask)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if got != nil {
			t.Fatal("expected nil return with error")
		}
		if err.Error() != c.Error {
			t.Errorf("got error %q; expect %q", err, c.Error)
		}
	}
}

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
		List  string
		Error string
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
		if err.Error() != c.Error {
			t.Errorf("got error %q; expect %q", err, c.Error)
		}
	}
}
