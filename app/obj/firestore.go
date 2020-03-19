package obj

import (
	"context"

	"cloud.google.com/go/firestore"
)

type firestorestore struct {
	client *firestore.Client
}

// NewFirestore builds an object store backed by Google Firestore.
func NewFirestore(c *firestore.Client) Store {
	return &firestorestore{
		client: c,
	}
}

func (s *firestorestore) Get(ctx context.Context, k Key, v Object) error {
	if err := validatekey(k); err != nil {
		return err
	}

	// Fetch from firestore.
	snap, err := s.client.Collection(k.Type()).Doc(k.ID()).Get(ctx)
	if err != nil {
		return err
	}

	// Unmarshal.
	return snap.DataTo(v)
}

func (s *firestorestore) Set(ctx context.Context, v Object) error {
	if err := validatekey(v); err != nil {
		return err
	}

	_, err := s.client.Collection(v.Type()).Doc(v.ID()).Set(ctx, v)
	return err
}
