package units

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestHumanize(t *testing.T) {
	cases := []struct {
		Input  Quantity
		Expect Quantity
	}{
		{
			Input:  Quantity{1_230_000, "ns/op"},
			Expect: Quantity{1.230, "ms/op"},
		},
		{
			Input:  Quantity{0.5, "MB/s"},
			Expect: Quantity{500, "KB/s"},
		},
		{
			Input:  Quantity{1234, "MB/s"},
			Expect: Quantity{1.234, "GB/s"},
		},
		{
			Input:  Quantity{1_234_000, "B/op"},
			Expect: Quantity{1.234, "MB/op"},
		},
		{
			Input:  Quantity{42.3, "widgets"},
			Expect: Quantity{42.3, "widgets"},
		},
	}
	for _, c := range cases {
		got := Humanize(c.Input)
		if diff := cmp.Diff(c.Expect, got); diff != "" {
			t.Errorf("mismatch\n%s", diff)
		}
	}
}

type TestCase struct {
	Value  float64
	Expect Quantity
}

func TableTest(t *testing.T, f func(float64) Quantity, cases []TestCase) {
	for _, c := range cases {
		got := f(c.Value)
		if diff := cmp.Diff(c.Expect, got); diff != "" {
			t.Errorf("mismatch\n%s", diff)
		}
	}
}

func TestDuration(t *testing.T) {
	TableTest(t, Duration, []TestCase{
		{13, Quantity{13, "ns"}},
		{1.23 * float64(time.Microsecond), Quantity{1.23, "\u00B5s"}},
		{1.23 * float64(time.Millisecond), Quantity{1.23, "ms"}},
		{1.23 * float64(time.Second), Quantity{1.23, "s"}},
		{1.23 * float64(time.Minute), Quantity{1.23, "m"}},
		{1.23 * float64(time.Hour), Quantity{1.23, "h"}},
	})
}

func TestBytesSI(t *testing.T) {
	TableTest(t, BytesSI, []TestCase{
		{13, Quantity{13, "B"}},
		{1_400, Quantity{1.4, "KB"}},
		{1_230_000, Quantity{1.23, "MB"}},
		{1_230_000_000, Quantity{1.23, "GB"}},
		{1_230_000_000_000, Quantity{1.23, "TB"}},
		{1_230_000_000_000_000, Quantity{1.23, "PB"}},
		{1_230_000_000_000_000_000, Quantity{1.23, "EB"}},
	})
}

func TestBytesBinary(t *testing.T) {
	TableTest(t, BytesBinary, []TestCase{
		{13, Quantity{13, "B"}},
		{0x2p10, Quantity{2, "KiB"}},
		{0x2p20, Quantity{2, "MiB"}},
		{0x2p30, Quantity{2, "GiB"}},
		{0x2p40, Quantity{2, "TiB"}},
		{0x2p50, Quantity{2, "PiB"}},
		{0x2p60, Quantity{2, "EiB"}},
	})
}

func TestScale(t *testing.T) {
	roman := func(v float64) Quantity {
		return scale(v, []Quantity{
			{1, "I"},
			{5, "V"},
			{10, "X"},
			{50, "L"},
			{100, "C"},
			{500, "D"},
			{1000, "M"},
		})
	}
	TableTest(t, roman, []TestCase{
		{3, Quantity{3, "I"}},
		{7, Quantity{1.4, "V"}},
		{40, Quantity{4, "X"}},
		{70, Quantity{1.4, "L"}},
		{200, Quantity{2, "C"}},
		{700, Quantity{1.4, "D"}},
		{1000, Quantity{1, "M"}},
		{1234, Quantity{1.234, "M"}},
	})
}
