// Package id provides helpers for repeatable ID generation.
package id

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// UUID generates a UUIDv5 from the given data.
func UUID(space uuid.UUID, data []byte) uuid.UUID {
	return uuid.NewSHA1(space, data)
}

// Strings produces a UUID for the given list of strings in the suppied name space.
func Strings(space uuid.UUID, s []string) uuid.UUID {
	return hash(space, s)
}

// KeyValues produces a UUID for the given key-value map in the suppied name space.
func KeyValues(space uuid.UUID, m map[string]string) uuid.UUID {
	return hash(space, m)
}

// hash any value by computing the UUID hash of the JSON-encoded value.
func hash(space uuid.UUID, v interface{}) uuid.UUID {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("converting value to json: %w", err))
	}
	return UUID(space, b)
}
