package obj

import (
	"context"
)

type overlay struct {
	a, b Store
}

// Overlay multiple object stores on top of each other, similar to a layer of
// caches. Gets will be issued to the stores in order, returning the first
// successful result. Stores will be issues to stores in order, stopping when a
// store already contains the item.
func Overlay(stores ...Store) Store {
	switch len(stores) {
	case 0:
		return Null
	case 1:
		return stores[0]
	default:
		return &overlay{
			a: stores[0],
			b: Overlay(stores[1:]...),
		}
	}
}

func (o *overlay) Contains(ctx context.Context, k Key) bool {
	if o.a.Contains(ctx, k) {
		return true
	}
	return o.b.Contains(ctx, k)
}

func (o *overlay) Get(ctx context.Context, k Key, v Object) error {
	// If the first store has it, return.
	if err := o.a.Get(ctx, k, v); err == nil {
		return nil
	}

	// Defer to the second.
	if err := o.b.Get(ctx, k, v); err != nil {
		return err
	}

	// Store in the first if found.
	return o.a.Set(ctx, v)
}

func (o *overlay) Set(ctx context.Context, v Object) error {
	// Nothing to be done if this layer contains it.
	if o.a.Contains(ctx, v) {
		return nil
	}

	// Store in first layer.
	if err := o.a.Set(ctx, v); err != nil {
		return err
	}

	// Delegate to layers below.
	return o.b.Set(ctx, v)
}
