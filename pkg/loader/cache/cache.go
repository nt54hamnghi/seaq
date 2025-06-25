package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/nt54hamnghi/seaq/pkg/config"
	"github.com/nt54hamnghi/seaq/pkg/env"
	"github.com/nt54hamnghi/seaq/pkg/util/log"
	"github.com/spf13/afero"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"go.etcd.io/bbolt"
)

const CacheFileName = "cache.db"

var fs = afero.Afero{
	Fs: afero.NewOsFs(),
}

// CacheLoader represents a documentloaders.Loader whose results can be cached.
type CacheableLoader interface {
	documentloaders.Loader
	// Hash returns a unique identifier for the loader, which is used as the cache key.
	// Implementations of this method must guarantee that if k1 == k2, then hash(k1) == hash(k2).
	Hash() ([]byte, error)
	// Type returns the loader's type which is used as the cache bucket.
	Type() string
}

// MarshalAndHash marshals a value to JSON and returns its xxHash as a hex string.
// It's used to generate consistent cache keys from structured data.
func MarshalAndHash(v any) ([]byte, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	h := xxhash.New()
	if n, err := h.Write(buf); err != nil {
		return nil, err
	} else if n != len(buf) {
		return nil, fmt.Errorf("failed to hash all bytes")
	}
	return h.Sum(nil), nil
}

// New creates a new CacheStorage.
func New(l CacheableLoader) (*Storage, error) {
	dir, _, err := config.AppConfig()
	if err != nil {
		return nil, err
	}

	if exists, err := fs.IsDir(dir); err != nil {
		return nil, err
	} else if !exists {
		if err := fs.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	return NewWithPath(l, filepath.Join(dir, CacheFileName))
}

// NewWithPath creates a new CacheStorage with a given path.
func NewWithPath(l CacheableLoader, path string) (*Storage, error) {
	db, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, err
	}

	return &Storage{Loader: l, db: db}, nil
}

type Storage struct {
	Loader CacheableLoader
	db     *bbolt.DB
}

type cacheItem struct {
	Docs      []schema.Document
	CreatedAt time.Time
}

func (ci cacheItem) expired() bool {
	return time.Since(ci.CreatedAt) > env.CacheDuration()
}

// id returns the bucket and key for the current loader.
// The bucket is derived from the loader's type, and the key from its hash.
// Returns an error if the loader's hash cannot be computed.
func (c Storage) id() (bucket []byte, key []byte, err error) {
	bucket = []byte(c.Loader.Type())
	if key, err = c.Loader.Hash(); err != nil {
		return nil, nil, err
	}
	return
}

func (c Storage) get() ([]schema.Document, error) {
	bucket, key, err := c.id()
	if err != nil {
		return nil, err
	}

	var raw []byte
	// read from cache
	err = c.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucket)
		// if bucket is not found, return nil,
		// raw is not modified and remains nil.
		if b == nil {
			return nil
		}

		// Get returns nil if key is not found,
		// so raw would also be nil.
		v := b.Get(key)
		raw = append(make([]byte, 0, len(v)), v...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// equivalent to if the neither bucket nor key is found
	if len(raw) == 0 {
		return nil, nil
	}

	var item cacheItem
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, err
	}

	// remove expired cache item
	if item.expired() {
		err = c.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(bucket)
			// bucket is not found, nothing to remove
			if b == nil {
				return nil
			}
			// Delete returns an error if it's called in a read-only transaction.
			// db.Update() creates a read-write transaction, so it's safe to ignore the error.
			_ = b.Delete(key)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to remove expired cache item: %w", err)
		}
		return nil, fmt.Errorf("cache expired")
	}

	return item.Docs, nil
}

func (c Storage) put(docs []schema.Document) error {
	bucket, key, err := c.id()
	if err != nil {
		return err
	}

	item := cacheItem{
		Docs:      docs,
		CreatedAt: time.Now(),
	}

	buf, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return c.db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}

		return b.Put(key, buf)
	})
}

func (c Storage) Close() error {
	return c.db.Close()
}

// Load loads from a source and returns documents.
func (c Storage) Load(ctx context.Context) ([]schema.Document, error) {
	docs, err := c.get()
	if err != nil {
		// failed to read cache is not fatal, we can still load from source
		log.Warn("failed to read cache", "error", err)
	}
	// if docs exist (cache hits)
	if docs != nil {
		return docs, nil
	}

	// if docs do not exist (cache misses), load from source
	docs, err = c.Loader.Load(ctx)
	if err != nil {
		return nil, err
	}

	// store docs to cache
	if err := c.put(docs); err != nil {
		// failed to write cache is not fatal, we can still return the docs
		log.Warn("failed to write cache", "error", err)
	}

	return docs, nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (c Storage) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := c.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
