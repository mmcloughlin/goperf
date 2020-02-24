package cpuset

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Set of unsigned integers.
type Set map[uint]bool

// NewSet builds a new set and inserts the given members into it.
func NewSet(members ...uint) Set {
	s := Set{}
	for _, member := range members {
		s[member] = true
	}
	return s
}

// Equals returns whether s and other are equal.
func (s Set) Equals(other Set) bool {
	if len(s) != len(other) {
		return false
	}
	for n := range s {
		if _, ok := other[n]; !ok {
			return false
		}
	}
	return true
}

// Reference: https://github.com/mkerrisk/man-pages/blob/ffea2c14f25042b1904e95da73d165cb25672a08/man7/cpuset.7#L907-L921
//
//	.\" ================== List Format ==================
//	.SS List format
//	The \fBList Format\fR for
//	.I cpus
//	and
//	.I mems
//	is a comma-separated list of CPU or memory-node
//	numbers and ranges of numbers, in ASCII decimal.
//	.PP
//	Examples of the \fBList Format\fR:
//	.PP
//	.in +4n
//	.EX
//	0\-4,9           # bits 0, 1, 2, 3, 4, and 9 set
//	0\-2,7,12\-14     # bits 0, 1, 2, 7, 12, 13, and 14 set
//

// FormatList represents s in "List Format".
func (s Set) FormatList() string {
	if len(s) == 0 {
		return ""
	}

	// Collect members in a sorted list.
	m := make([]uint, 0, len(s))
	for i := range s {
		m = append(m, i)
	}
	sort.Slice(m, func(i, j int) bool { return m[i] < m[j] })

	// Build list
	list := ""
	for len(m) > 0 {
		// Determine length of run.
		n := 0
		for ; n < len(m) && m[n] == m[0]+uint(n); n++ {
		}
		// Append to list.
		if list != "" {
			list += ","
		}
		list += strconv.FormatUint(uint64(m[0]), 10)
		if n > 1 {
			list += "-" + strconv.FormatUint(uint64(m[n-1]), 10)
		}
		// Advance.
		m = m[n:]
	}

	return list
}

// ParseList parses a set represented in "List Format", specifically a
// comma-separated list of ranges.
func ParseList(s string) (Set, error) {
	if s == "" {
		return NewSet(), nil
	}
	ranges := strings.Split(s, ",")
	set := NewSet()
	for _, r := range ranges {
		start, end, err := parserange(r)
		if err != nil {
			return nil, err
		}
		for i := start; i <= end; i++ {
			set[i] = true
		}
	}
	return set, nil
}

func parserange(s string) (uint, uint, error) {
	parts := strings.Split(s, "-")
	switch len(parts) {
	case 1:
		n, err := parseuint(parts[0])
		return n, n, err

	case 2:
		start, err := parseuint(parts[0])
		if err != nil {
			return 0, 0, err
		}
		end, err := parseuint(parts[1])
		if err != nil {
			return 0, 0, err
		}
		if start > end {
			return 0, 0, fmt.Errorf("start exceeds end in range %q", s)
		}
		return start, end, nil

	default:
		return 0, 0, fmt.Errorf("range %q has too many splits", s)
	}
}

func parseuint(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse unsigned integer: %q invalid", s)
	}
	return uint(n), nil
}
