package cpuset_test

import (
	"fmt"

	"github.com/mmcloughlin/cb/pkg/cpuset"
)

func ExampleSet_FormatMask() {
	s := cpuset.NewSet(32, 64, 0, 1, 4, 2, 8, 16)
	fmt.Println(s.FormatMask())
	// Output:
	// 00000001,00000001,00010117
}

func ExampleParseMask() {
	s, _ := cpuset.ParseMask("40000000,00000001,00000000")
	for n := range s {
		fmt.Println(n)
	}
	// Unordered output:
	// 32
	// 94
}

func ExampleSet_FormatList() {
	s := cpuset.NewSet(13, 2, 0, 7, 1, 14, 12)
	fmt.Println(s.FormatList())
	// Output:
	// 0-2,7,12-14
}

func ExampleParseList() {
	s, _ := cpuset.ParseList("0-2,7,12-14")
	for n := range s {
		fmt.Println(n)
	}
	// Unordered output:
	// 0
	// 1
	// 2
	// 7
	// 12
	// 13
	// 14
}
