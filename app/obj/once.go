package obj

import "context"

type once struct {
	s       Setter
	written map[Key]bool
}

// OnceSetter wraps a setter s to ensure that each key will only be set once.
// Note that the underlying set of written keys is unbounded.
func OnceSetter(s Setter) Setter {
	return &once{
		s:       s,
		written: map[Key]bool{},
	}
}

func (o *once) Set(ctx context.Context, v Object) error {
	k := K(v.Type(), v.ID())
	if o.written[k] {
		return nil
	}

	if err := o.s.Set(ctx, v); err != nil {
		return err
	}

	o.written[k] = true
	return nil
}
