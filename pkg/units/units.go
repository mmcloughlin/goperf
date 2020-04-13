// Package units implements human-friendly representations of common units.
package units

import "time"

// Standard library testing package units.
const (
	Runtime        = "ns/op"
	DataRate       = "MB/s"
	BytesAllocated = "B/op"
	Allocs         = "allocs/op"
)

// Quantity is a value in some unit.
type Quantity struct {
	Value float64
	Unit  string
}

// Humanize attempts to represent q in a more friendly unit for human
// consumption.
func Humanize(q Quantity) Quantity {
	switch q.Unit {
	case Runtime:
		q = Duration(q.Value)
		q.Unit += "/op"
	case DataRate:
		q = BytesSI(q.Value * 1e6)
		q.Unit += "/s"
	case BytesAllocated:
		q = BytesSI(q.Value)
		q.Unit += "/op"
	}
	return q
}

// Duration represents a nanosecond quantity in time units up to hours.
func Duration(ns float64) Quantity {
	return scale(ns, []Quantity{
		{float64(time.Nanosecond), "ns"},
		{float64(time.Microsecond), "\u00B5s"}, // https://decodeunicode.org/U+00B5
		{float64(time.Millisecond), "ms"},
		{float64(time.Second), "s"},
		{float64(time.Minute), "m"},
		{float64(time.Hour), "h"},
	})
}

// BytesSI represents the given number of bytes with SI units (multiples of
// 1000).
func BytesSI(b float64) Quantity {
	return scale(b, []Quantity{
		{1, "B"},
		{1e3, "KB"},
		{1e6, "MB"},
		{1e9, "GB"},
		{1e12, "TB"},
		{1e15, "PB"},
		{1e18, "EB"},
	})
}

// BytesBinary represents the given number of bytes with binary prefixes
// (multiples of 1024).
func BytesBinary(b float64) Quantity {
	return scale(b, []Quantity{
		{1, "B"},
		{0x1p10, "KiB"},
		{0x1p20, "MiB"},
		{0x1p30, "GiB"},
		{0x1p40, "TiB"},
		{0x1p50, "PiB"},
		{0x1p60, "EiB"},
	})
}

// Frequency represents a frequency in Hertz.
func Frequency(f float64) Quantity {
	return scale(f, []Quantity{
		{1, "Hz"},
		{1e3, "KHz"},
		{1e6, "MHz"},
		{1e9, "GHz"},
	})
}

func scale(v float64, units []Quantity) Quantity {
	i := 0
	for i+1 < len(units) && v >= units[i+1].Value {
		i++
	}
	return Quantity{
		Value: v / units[i].Value,
		Unit:  units[i].Unit,
	}
}
