package cpuset_test

import (
	"fmt"

	"github.com/mmcloughlin/cb/pkg/cpuset"
)

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
