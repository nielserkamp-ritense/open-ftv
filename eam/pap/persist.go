package pap

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// PathSeparator is the standard separator character to use with multi-level keys.
const PathSeparator = "/"

// Persistence represents the interface to manage persistent storage for policies.
type Persistence interface {
	Create(p Policy) (Policy, error)
	Read(language, id string) (Policy, uint64, error)
	Update(prev Policy, lastIndex uint64, p Policy) (Policy, error)
	Delete(prev Policy, lastIndex uint64) (Policy, error)
	List(language string) ([]Policy, error)
	// ReadByHash resolves a policy by the SHA-256 of its source content, from the
	// content-addressable index maintained on every Create/Update.
	ReadByHash(hash string) (Policy, error)
}

// NewStore instantiates a new persistent storage handler for policies.
//
// It creates a CRUD wrapper around the given Valkeyrie Store interface.
// This means it can be used with various distributed KV backends,
// such as Consul, Etcd, Zookeeper, BoltDB, DynamoDB.
// See https://github.com/kvtools/valkeyrie.
//
// Use basePath to define the key prefix to use for the backend KV store.
// All policy id's will be prefixed with the value of basePath as the KV store's key.
// A trailing PathSeparator character in basePath is automatically appended if it is missing from the input.
//
// The given context is passed in every call to the KV backend.
func NewStore(ctx context.Context, client store.Store, basePath string) Persistence {
	return &wrapper{ctx: ctx, client: client, basePath: convert.ForceSuffix(basePath, PathSeparator)}
}

// Create implements the Persistence interface.
func (s *wrapper) Create(p Policy) (Policy, error) {
	key := s.makeKey(p.Language(), p.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, _, err := s.client.AtomicPut(s.ctx, key, s.mustMarshal(p), nil, writeOptions); err != nil {
		return s.failure("create", p.ID(), err, false)
	}
	s.indexByHash(p)
	return p, nil
}

// Read implements the Persistence interface.
func (s *wrapper) Read(language, id string) (Policy, uint64, error) {
	key := s.bugFix(s.makeKey(language, id))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	kv, err := s.client.Get(s.ctx, key, readOptions)
	if err != nil || kv == nil {
		p, err2 := s.failure("read", id, err, true)
		return p, 0, err2
	}

	p, err2 := s.unmarshal(id, kv)
	if err2 != nil {
		p, err2 = s.failure("read", id, err2, true)
		return p, 0, err2
	}

	return p, kv.LastIndex, nil
}

// Update implements the Persistence interface.
func (s *wrapper) Update(prev Policy, lastIndex uint64, p Policy) (Policy, error) {
	key := s.makeKey(prev.Language(), prev.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: s.mustMarshal(prev), LastIndex: lastIndex}
	if _, _, err := s.client.AtomicPut(s.ctx, key, s.mustMarshal(p), &kv, writeOptions); err != nil {
		return s.failure("update", prev.ID(), err, true)
	}
	s.indexByHash(p)
	return p, nil
}

// Delete implements the Persistence interface.
func (s *wrapper) Delete(prev Policy, lastIndex uint64) (Policy, error) {
	key := s.makeKey(prev.Language(), prev.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: s.mustMarshal(prev), LastIndex: lastIndex}
	if _, err := s.client.AtomicDelete(s.ctx, key, &kv); err != nil {
		return s.failure("delete", prev.ID(), err, true)
	}

	return prev, nil
}

// List implements the Persistence interface.
func (s *wrapper) List(language string) ([]Policy, error) {
	key := s.basePath
	if language != "" {
		key = fmt.Sprintf("%s%s%s", key, language, PathSeparator)
	}
	key = s.bugFix(key)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list, err := s.client.List(s.ctx, key, readOptions)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read policies: %w", err)
	}

	out := make([]Policy, 0, len(list))
	for _, kv := range list {
		if s.isHashIndexKey(kv.Key) {
			// content-addressable index entry, not a distinct policy: skip it so the
			// index never shows up as duplicate policies in List (including List("")).
			continue
		}
		p := new(policy)
		if err = json.Unmarshal(kv.Value, p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal policies: %w", err)
		}
		out = append(out, p)
	}

	return out, nil
}

// ReadByHash implements the Persistence interface: it resolves a policy by the
// SHA-256 of its source content, using the content-addressable index. Older
// versions of a policy remain retrievable under their own hash even after the
// live policy has been updated - callers depending on this for replay must keep in
// mind the retention implication: the index grows monotonically (one entry per
// distinct content ever stored) and is never garbage-collected here.
func (s *wrapper) ReadByHash(hash string) (Policy, error) {
	key := s.bugFix(s.basePath + hashPrefix + hash)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	kv, err := s.client.Get(s.ctx, key, readOptions)
	if err != nil || kv == nil {
		return nil, fmt.Errorf("policy for hash '%s' not found", hash)
	}
	return s.unmarshal(hash, kv)
}

// indexByHash writes (or overwrites, idempotently) the content-addressable index
// entry for a policy: key "hash/<sha256-of-content>" -> the marshalled policy. A
// failure to index is logged-through as a no-op: the primary write already
// succeeded, and re-indexing is deterministic, so it never fails the Create/Update.
func (s *wrapper) indexByHash(p Policy) {
	h, err := policyContentHash(p)
	if err != nil {
		return
	}
	key := s.bugFix(s.basePath + hashPrefix + h)
	_ = s.client.Put(s.ctx, key, s.mustMarshal(p), writeOptions)
}

// isHashIndexKey reports whether a raw store key belongs to the content-addressable
// index (pseudo-language "hash") rather than being a real "<language>/<id>" policy.
func (s *wrapper) isHashIndexKey(key string) bool {
	k := strings.TrimPrefix(key, "/")
	base := strings.TrimPrefix(s.basePath, "/")
	rel := strings.TrimPrefix(k, base)
	return strings.HasPrefix(rel, hashPrefix)
}

func (s *wrapper) makeKey(language, id string) string {
	return fmt.Sprintf("%s%s%s%s", s.basePath, language, PathSeparator, id)
}

func (s *wrapper) mustMarshal(p Policy) []byte {
	b, _ := json.Marshal(p)
	return b
}

func (s *wrapper) unmarshal(id string, kv *store.KVPair) (Policy, error) {
	p := new(policy)
	if err := json.Unmarshal(kv.Value, p); err != nil {
		return s.failure("unmarshal", id, err, true)
	}
	return p, nil
}

func (s *wrapper) failure(op, id string, err error, mustFind bool) (Policy, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to %s policy '%s': %w", op, id, err)
	}
	if mustFind {
		return nil, fmt.Errorf("policy '%s' not found", id)
	}
	return nil, fmt.Errorf("policy '%s' already exists", id)
}

func (s *wrapper) bugFix(in string) string {
	// the Valkeyrie/etcdv3 implementation sometimes removes a leading slash character from the key.
	if _, ok := s.client.(*etcdv3.Store); ok {
		return "/" + in
	}
	return in
}

type wrapper struct {
	ctx       context.Context
	client    store.Store
	basePath  string
	mutex     sync.RWMutex
	storeLock store.Locker
}

var (
	readOptions  = &store.ReadOptions{}
	writeOptions = &store.WriteOptions{}
)
