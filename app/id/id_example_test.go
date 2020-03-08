package id_test

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/mmcloughlin/cb/app/id"
)

func ExampleUUID() {
	space := uuid.MustParse("19433b12-5b85-4252-9bb0-594a38156713")
	ident := id.UUID(space, []byte("hello world"))
	fmt.Println(ident)
	// Output: b7ca611e-7a4d-59b9-9ea6-4aabd08ff088
}

func ExampleStrings() {
	space := uuid.MustParse("19433b12-5b85-4252-9bb0-594a38156713")
	ident := id.Strings(space, []string{"hello", "world"})
	fmt.Println(ident)
	// Output: 8cf4c7fd-c0a3-5134-90a5-531830a1c3e0
}

func ExampleKeyValues() {
	space := uuid.MustParse("19433b12-5b85-4252-9bb0-594a38156713")
	ident := id.KeyValues(space, map[string]string{"greeting": "hello", "who": "world"})
	fmt.Println(ident)
	// Output: 60376fde-597e-5694-8a19-e6dd930fbec0
}
