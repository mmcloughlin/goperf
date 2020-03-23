// Package obj provides abstractions for simple object storage.
package obj

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/vmihailenco/msgpack"
)

// KeyNotFoundError indicates that a key was not found in an object store.
type KeyNotFoundError struct {
	Key Key
}

// Error returns an error message, as required by the error interface.
func (e KeyNotFoundError) Error() string {
	k := e.Key
	return fmt.Sprintf("key (%s,%s) not found", k.Type(), k.ID())
}

// Key identifies an object.
type Key interface {
	Type() string
	ID() string
}

func validatekey(k Key) error {
	if k.Type() == "" {
		return errors.New("empty type")
	}
	if k.ID() == "" {
		return errors.New("empty id")
	}
	return nil
}

type key struct {
	t  string
	id string
}

// K builds a key with the given type and ID.
func K(t, id string) Key {
	return key{
		t:  t,
		id: id,
	}
}

func (k key) Type() string { return k.t }
func (k key) ID() string   { return k.id }

// Object is a serializable object with an associated Key. All objects should be
// JSON serializable. Other serializaiton methods may be required by certain
// storage systems.
type Object interface {
	Key
}

// Getter can get object by key.
type Getter interface {
	Contains(context.Context, Key) bool
	Get(context.Context, Key, Object) error
}

// Setter can write objects.
type Setter interface {
	Set(context.Context, Object) error
}

// Store is a method of storing objects by key.
type Store interface {
	Getter
	Setter
}

// Null contains nothing and stores nothing.
var Null Store = null{}

type null struct{}

func (null) Contains(context.Context, Key) bool           { return false }
func (null) Get(_ context.Context, k Key, v Object) error { return KeyNotFoundError{Key: k} }
func (null) Set(context.Context, Object) error            { return nil }

// Encode the object to w.
func Encode(w io.Writer, v Object) error {
	return msgpack.NewEncoder(w).UseJSONTag(true).Encode(v)
}

// Decode an object from r.
func Decode(r io.Reader, v Object) error {
	return msgpack.NewDecoder(r).UseJSONTag(true).Decode(v)
}
