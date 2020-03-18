// Package obj provides abstractions for simple object storage.
package obj

import (
	"context"
	"fmt"
	"io"

	"github.com/vmihailenco/msgpack"
)

type KeyNotFoundError struct {
	Key Key
}

func (e KeyNotFoundError) Error() string {
	k := e.Key
	return fmt.Sprintf("key (%s,%s) not found", k.Type(), k.ID())
}

type Key interface {
	Type() string
	ID() string
}

type key struct {
	t  string
	id string
}

func K(t, id string) Key {
	return key{
		t:  t,
		id: id,
	}
}

func (k key) Type() string { return k.t }
func (k key) ID() string   { return k.id }

type Object interface {
	Key
}

type Store interface {
	Get(context.Context, Key, Object) error
	Set(context.Context, Object) error
}

// Encode the object to w.
func Encode(w io.Writer, v Object) error {
	return msgpack.NewEncoder(w).UseJSONTag(true).Encode(v)
}

// Decode an object from r.
func Decode(r io.Reader, v Object) error {
	return msgpack.NewDecoder(r).UseJSONTag(true).Decode(v)
}
