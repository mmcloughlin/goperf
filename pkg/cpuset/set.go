package cpuset

import (
	"fmt"
	"math/big"
	"math/bits"
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

func (s Set) String() string {
	return s.FormatList()
}

// Equals returns whether s and t are equal.
func (s Set) Equals(t Set) bool {
	if len(s) != len(t) {
		return false
	}
	return s.Contains(t)
}

// Contains reports whether s contains t.
func (s Set) Contains(t Set) bool {
	if len(t) > len(s) {
		return false
	}
	for n := range t {
		if _, ok := s[n]; !ok {
			return false
		}
	}
	return true
}

// Members returns the members of s as a slice. No guarantees are made about order.
func (s Set) Members() []uint {
	m := make([]uint, 0, len(s))
	for n := range s {
		m = append(m, n)
	}
	return m
}

// SortedMembers returns the members of s as a slice in sorted order.
func (s Set) SortedMembers() []uint {
	m := s.Members()
	sort.Slice(m, func(i, j int) bool { return m[i] < m[j] })
	return m
}

// Clone returns a copy of s.
func (s Set) Clone() Set {
	t := NewSet()
	for n := range s {
		t[n] = true
	}
	return t
}

// Difference returns a new set with elements in s but not in t.
func (s Set) Difference(t Set) Set {
	d := s.Clone()
	for n := range t {
		delete(d, n)
	}
	return d
}

// Reference: https://github.com/mkerrisk/man-pages/blob/ffea2c14f25042b1904e95da73d165cb25672a08/man7/cpuset.7#L866-L906
//
//	.\" ================== Mask Format ==================
//	.SS Mask format
//	The \fBMask Format\fR is used to represent CPU and memory-node bit masks
//	in the
//	.I /proc/<pid>/status
//	file.
//	.PP
//	This format displays each 32-bit
//	word in hexadecimal (using ASCII characters "0" - "9" and "a" - "f");
//	words are filled with leading zeros, if required.
//	For masks longer than one word, a comma separator is used between words.
//	Words are displayed in big-endian
//	order, which has the most significant bit first.
//	The hex digits within a word are also in big-endian order.
//	.PP
//	The number of 32-bit words displayed is the minimum number needed to
//	display all bits of the bit mask, based on the size of the bit mask.
//	.PP
//	Examples of the \fBMask Format\fR:
//	.PP
//	.in +4n
//	.EX
//	00000001                        # just bit 0 set
//	40000000,00000000,00000000      # just bit 94 set
//	00000001,00000000,00000000      # just bit 64 set
//	000000ff,00000000               # bits 32\-39 set
//	00000000,000e3862               # 1,5,6,11\-13,17\-19 set
//	.EE
//	.in
//	.PP
//	A mask with bits 0, 1, 2, 4, 8, 16, 32, and 64 set displays as:
//	.PP
//	.in +4n
//	.EX
//	00000001,00000001,00010117
//	.EE
//	.in
//	.PP
//	The first "1" is for bit 64, the
//	second for bit 32, the third for bit 16, the fourth for bit 8, the
//	fifth for bit 4, and the "7" is for bits 2, 1, and 0.
//

// FormatMask represents s in linux "Mask Format".
func (s Set) FormatMask() string {
	// Build bitset.
	bitset := new(big.Int)
	for m := range s {
		bitset.SetBit(bitset, int(m), 1)
	}

	// Mask of low 32 bits.
	one := big.NewInt(1)
	low32 := new(big.Int).Lsh(one, 32)
	low32.Sub(low32, one)

	// Produce 32-bits at a time.
	mask := ""
	first := true
	zero := new(big.Int)
	for first || bitset.Cmp(zero) != 0 {
		word := new(big.Int).And(bitset, low32)
		mask = fmt.Sprintf(",%08x%s", word, mask)
		bitset.Rsh(bitset, 32)
		first = false
	}

	return mask[1:]
}

// ParseMask parses a set represented in linux "Mask Format", specifically a
// comma-separated list of 32-bit hex words in big-endian order.
func ParseMask(s string) (Set, error) {
	words := strings.Split(s, ",")
	l := uint(0)
	set := NewSet()
	for i := len(words) - 1; i >= 0; i-- {
		word := words[i]
		if len(word) != 8 {
			return nil, fmt.Errorf("parsing mask word %q: expected 8 hex characters", word)
		}
		x, err := strconv.ParseUint(word, 16, 32)
		if err != nil {
			return nil, fmt.Errorf("parsing mask word %q: invalid", word)
		}
		for b := l; x != 0; {
			n := uint(bits.TrailingZeros64(x))
			set[b+n] = true
			x >>= n + 1
			b += n + 1
		}
		l += 32
	}
	return set, nil
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
	m := s.SortedMembers()

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
