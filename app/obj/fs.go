package obj

import (
	"bytes"
	"context"
	"path"

	"github.com/golang/groupcache/lru"

	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/fs"
)

type filesystem struct {
	fs       fs.Interface
	size     int64
	capacity int64
	lru      *lru.Cache
}

type entry struct {
	path string
	size int64
}

// NewFileSystem builds an object store backed by the supplied filesystem that
// will store at most capacity bytes.
func NewFileSystem(fs fs.Interface, capacity int64) Store {
	return &filesystem{
		fs:       fs,
		size:     0,
		capacity: capacity,
		lru:      lru.New(0),
	}
}

func (f *filesystem) Get(ctx context.Context, k Key, v Object) error {
	// Lookup the key in the cache first.
	key := cachekey(k)
	value, ok := f.lru.Get(key)
	if !ok {
		return KeyNotFoundError{k}
	}
	e := value.(entry)

	// Read and decode the object.
	r, err := f.fs.Open(ctx, e.path)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := Decode(r, v); err != nil {
		return err
	}

	return nil
}

func (f *filesystem) Set(ctx context.Context, v Object) error {
	// Encode.
	buf := bytes.NewBuffer(nil)
	if err := Encode(buf, v); err != nil {
		return err
	}
	size := int64(buf.Len())

	// Make room for it.
	for f.size+size > f.capacity {
		if err := f.removeoldest(ctx); err != nil {
			return err
		}
	}

	// Write to the filesystem.
	path := keypath(v)
	if err := fs.WriteFile(ctx, f.fs, path, buf.Bytes()); err != nil {
		return err
	}

	// Add to the cache.
	key := cachekey(v)
	f.lru.Add(key, entry{
		path: path,
		size: size,
	})

	// Bump total storage.
	f.size += size

	return nil
}

func (f *filesystem) removeoldest(ctx context.Context) error {
	var err error
	f.lru.OnEvicted = func(_ lru.Key, v interface{}) {
		err = f.evict(ctx, v.(entry))
	}
	f.lru.RemoveOldest()
	f.lru.OnEvicted = nil
	return err
}

func (f *filesystem) evict(ctx context.Context, e entry) error {
	// Delete from the filesystem.
	if err := f.fs.Remove(ctx, e.path); err != nil {
		return err
	}

	// Free up space.
	f.size -= e.size

	if f.size < 0 {
		return errutil.AssertionFailure("negative size")
	}

	return nil
}

// keypath returns the filepath for the object with the given key.
func keypath(k Key) string {
	return path.Join(k.Type(), k.ID())
}

// cachekey returns the key in the LRU for object key k.
func cachekey(k Key) lru.Key {
	return key{t: k.Type(), id: k.ID()}
}
