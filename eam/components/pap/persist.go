package pap

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/kvtools/valkeyrie/store"
)

const pathSeparator = "/"

// Persistence represents the interface to manage persistent storage for policies.
type Persistence interface {
	Create(p Policy) (Policy, error)
	Read(language, id string) (Policy, error)
	Update(prev, p Policy) (Policy, error)
	Delete(prev Policy) (Policy, error)
	List(language string) ([]Policy, error)
}

// NewStore instantiates a new persistent storage handler for policies.
//
// It creates a CRUD wrapper around the given Valkeyrie Store interface.
// This means it can be used with various distributed KV backends,
// such as Consul, etcd, Zookeeper, BoltDB, DynamoDB.
// See https://github.com/kvtools/valkeyrie.
//
// Use basePath to define the key prefix to use for the backend KV store.
// All policy id's will be prefixed with the value of basePath as the KV store's key.
// A trailing pathSeparator character in basePath is automatically appended if it is missing from the input.
//
// The given context is passed in every call to the KV backend.
func NewStore(ctx context.Context, client store.Store, basePath string) Persistence {
	_, err := client.NewLock(ctx, "", &store.LockOptions{})
	canLock := !errors.Is(err, store.ErrCallNotSupported)

	if basePath != "" && !strings.HasSuffix(basePath, pathSeparator) {
		basePath += pathSeparator
	}

	return &wrapper{ctx: ctx, client: client, canLock: canLock, basePath: basePath}
}

// Create implements the Persistence interface.
func (s *wrapper) Create(p Policy) (Policy, error) {
	key := s.makeKey(p.Language(), p.ID())

	if err := s.lock(key); err != nil {
		return s.failure("create", p.ID(), err, false)
	}
	defer s.unlock()

	if _, _, err := s.client.AtomicPut(s.ctx, key, s.mustMarshal(p), nil, writeOptions); err != nil {
		return s.failure("create", p.ID(), err, false)
	}
	return p, nil
}

// Read implements the Persistence interface.
func (s *wrapper) Read(language, id string) (Policy, error) {
	key := s.makeKey(language, id)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	kv, err := s.client.Get(s.ctx, key, readOptions)
	if err != nil || kv == nil {
		return s.failure("read", id, err, true)
	}

	p, err2 := s.unmarshal(id, kv)
	if err2 != nil {
		return s.failure("read", id, err2, true)
	}
	return p, nil
}

// Update implements the Persistence interface.
func (s *wrapper) Update(prev, p Policy) (Policy, error) {
	key := s.makeKey(prev.Language(), prev.ID())

	if err := s.lock(key); err != nil {
		return s.failure("update", prev.ID(), err, true)
	}
	defer s.unlock()

	kv := store.KVPair{Key: key, Value: s.mustMarshal(prev)}
	if _, _, err := s.client.AtomicPut(s.ctx, key, s.mustMarshal(p), &kv, writeOptions); err != nil {
		return s.failure("update", prev.ID(), err, true)
	}
	return p, nil
}

// Delete implements the Persistence interface.
func (s *wrapper) Delete(prev Policy) (Policy, error) {
	key := s.makeKey(prev.Language(), prev.ID())

	if err := s.lock(key); err != nil {
		return s.failure("delete", prev.ID(), err, true)
	}
	defer s.unlock()

	kv := store.KVPair{Key: key, Value: s.mustMarshal(prev)}
	if _, err := s.client.AtomicDelete(s.ctx, key, &kv); err != nil {
		return s.failure("delete", prev.ID(), err, true)
	}

	return prev, nil
}

// List implements the Persistence interface.
func (s *wrapper) List(language string) ([]Policy, error) {
	key := s.basePath
	if language != "" {
		key = fmt.Sprintf("%s%s%s", key, language, pathSeparator)
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list, err := s.client.List(s.ctx, key, readOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to read policies: %w", err)
	}

	out := make([]Policy, 0, len(list))
	for _, kv := range list {
		p := new(policy)
		if err = json.Unmarshal(kv.Value, p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal policies: %w", err)
		}
		out = append(out, p)
	}

	return out, nil
}

func (s *wrapper) makeKey(language, id string) string {
	return fmt.Sprintf("%s%s%s%s", s.basePath, language, pathSeparator, id)
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

func (s *wrapper) lock(key string) error {
	s.mutex.Lock()
	if !s.canLock {
		return nil
	}

	lock, err := s.client.NewLock(s.ctx, key, lockOptions)
	if err != nil {
		s.mutex.Unlock()
		return err
	}

	_, err = lock.Lock(s.ctx)
	if err != nil {
		s.mutex.Unlock()
		return err
	}

	s.storeLock = lock
	return nil
}

func (s *wrapper) unlock() {
	if s.storeLock != nil {
		s.storeLock.Unlock(s.ctx)
		s.storeLock = nil
	}
	s.mutex.Unlock()
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

type wrapper struct {
	ctx       context.Context
	client    store.Store
	canLock   bool
	basePath  string
	mutex     sync.RWMutex
	storeLock store.Locker
}

var (
	readOptions  = &store.ReadOptions{Consistent: true}
	writeOptions = &store.WriteOptions{}
	lockOptions  = &store.LockOptions{}
)
