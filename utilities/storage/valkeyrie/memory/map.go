// Package memory implements a memory-mapped Valkeyrie KV-store.
package memory

import (
	"bytes"
	"context"
	"strings"
	"sync"

	"github.com/kvtools/valkeyrie/store"
)

// New instantiates a memory mapped Valkeyrie Store.
func New() store.Store {
	return &kv{kv: make(map[string][]byte, 64)}
}

// Put implements the Valkeyrie.Store interface.
func (k *kv) Put(_ context.Context, key string, value []byte, _ *store.WriteOptions) error {
	k.mutex.Lock()
	k.kv[key] = value
	k.mutex.Unlock()
	return nil
}

// Get implements the Valkeyrie.Store interface.
func (k *kv) Get(_ context.Context, key string, _ *store.ReadOptions) (*store.KVPair, error) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	if v := k.kv[key]; v != nil {
		return &store.KVPair{Key: key, Value: v}, nil
	}
	return nil, store.ErrKeyNotFound
}

// Delete implements the Valkeyrie.Store interface.
func (k *kv) Delete(_ context.Context, key string) error {
	k.mutex.Lock()
	delete(k.kv, key)
	k.mutex.Unlock()
	return nil
}

// Exists implements the Valkeyrie.Store interface.
func (k *kv) Exists(_ context.Context, key string, _ *store.ReadOptions) (bool, error) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	if v := k.kv[key]; v != nil {
		return true, nil
	}
	return false, store.ErrKeyNotFound
}

// Watch implements the Valkeyrie.Store interface.
func (k *kv) Watch(_ context.Context, _ string, _ *store.ReadOptions) (<-chan *store.KVPair, error) {
	return nil, store.ErrCallNotSupported
}

// WatchTree implements the Valkeyrie.Store interface.
func (k *kv) WatchTree(_ context.Context, _ string, _ *store.ReadOptions) (<-chan []*store.KVPair, error) {
	return nil, store.ErrCallNotSupported
}

// NewLock implements the Valkeyrie.Store interface.
func (k *kv) NewLock(_ context.Context, _ string, _ *store.LockOptions) (store.Locker, error) {
	return nil, store.ErrCallNotSupported
}

// List implements the Valkeyrie.Store interface.
func (k *kv) List(_ context.Context, directory string, _ *store.ReadOptions) ([]*store.KVPair, error) {
	if directory != "" && !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	k.mutex.RLock()

	out := make([]*store.KVPair, 0)
	for key := range k.kv {
		if directory == "" || strings.HasPrefix(key, directory) {
			out = append(out, &store.KVPair{Key: key, Value: k.kv[key]})
		}
	}

	k.mutex.RUnlock()
	return out, nil
}

// DeleteTree implements the Valkeyrie.Store interface.
func (k *kv) DeleteTree(_ context.Context, directory string) error {
	if directory != "" && !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	k.mutex.Lock()

	for key := range k.kv {
		if directory == "" || strings.HasPrefix(key, directory) {
			delete(k.kv, key)
		}
	}

	k.mutex.Unlock()
	return nil
}

// AtomicPut implements the Valkeyrie.Store interface.
func (k *kv) AtomicPut(_ context.Context, key string, value []byte, previous *store.KVPair, _ *store.WriteOptions) (bool, *store.KVPair, error) {
	if previous != nil && key != previous.Key {
		return false, nil, store.ErrKeyModified
	}

	k.mutex.Lock()

	v, ok := k.kv[key]
	if previous != nil {
		if ok && bytes.Equal(v, previous.Value) {
			k.kv[key] = value
		}
	} else {
		if !ok {
			k.kv[key] = value
		}
	}

	k.mutex.Unlock()

	switch {
	case previous == nil && ok:
		return false, nil, store.ErrKeyExists
	case previous != nil && !ok:
		return false, nil, store.ErrKeyNotFound
	case previous != nil && !bytes.Equal(v, previous.Value):
		return false, nil, store.ErrKeyModified
	default:
		return true, &store.KVPair{Key: key, Value: value}, nil
	}
}

// AtomicDelete implements the Valkeyrie.Store interface.
func (k *kv) AtomicDelete(_ context.Context, key string, previous *store.KVPair) (bool, error) {
	k.mutex.Lock()

	v, ok := k.kv[key]
	if ok && bytes.Equal(v, previous.Value) {
		delete(k.kv, key)
	}

	k.mutex.Unlock()

	switch {
	case !ok:
		return false, store.ErrKeyNotFound
	case !bytes.Equal(v, previous.Value):
		return false, store.ErrKeyModified
	default:
		return true, nil
	}
}

// Close implements the Valkeyrie.Store interface.
func (k *kv) Close() error {
	// no-op
	return nil
}

type kv struct {
	kv    map[string][]byte
	mutex sync.RWMutex
}
