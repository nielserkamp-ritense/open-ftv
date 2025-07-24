package pap

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/goccy/go-json"
	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// PathSeparator is the standard separator character to use with multi-level keys.
const PathSeparator = "/"

// NewPersistence instantiates a new persistent storage handler for policies.
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
func NewPersistence(ctx context.Context, client store.Store, basePath string) *Persistence {
	return &Persistence{ctx: ctx, client: client, basePath: convert.ForceSuffix(basePath, PathSeparator)}
}

// Persistence contains the details to manage persistent storage for policies.
type Persistence struct {
	ctx      context.Context
	client   store.Store
	basePath string
	mutex    sync.RWMutex
}

// Create adds a policy to the store.
func (s *Persistence) Create(p *models.Policy) (*models.Policy, error) {
	key := s.makeKey(p.Language(), p.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, _, err := s.client.AtomicPut(s.ctx, key, s.mustMarshal(p), nil, writeOptions); err != nil {
		return s.failure("create", p.ID(), err, false)
	}
	return p, nil
}

// Read retrieves a policy from the store.
func (s *Persistence) Read(language, id string) (*models.Policy, uint64, error) {
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

// Update replaces a policy in the store.
func (s *Persistence) Update(prev *models.Policy, lastIndex uint64, p *models.Policy) (*models.Policy, error) {
	key := s.makeKey(prev.Language(), prev.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: s.mustMarshal(prev), LastIndex: lastIndex}
	if _, _, err := s.client.AtomicPut(s.ctx, key, s.mustMarshal(p), &kv, writeOptions); err != nil {
		return s.failure("update", prev.ID(), err, true)
	}
	return p, nil
}

// Delete removes a policy from the store.
func (s *Persistence) Delete(prev *models.Policy, lastIndex uint64) (*models.Policy, error) {
	key := s.makeKey(prev.Language(), prev.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: s.mustMarshal(prev), LastIndex: lastIndex}
	if _, err := s.client.AtomicDelete(s.ctx, key, &kv); err != nil {
		return s.failure("delete", prev.ID(), err, true)
	}

	return prev, nil
}

// List returns all policies from the store.
func (s *Persistence) List(language string) ([]*models.Policy, error) {
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

	out := make([]*models.Policy, 0, len(list))
	for _, kv := range list {
		p := new(models.Policy)
		if err = json.Unmarshal(kv.Value, p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal policies: %w", err)
		}
		out = append(out, p)
	}

	return out, nil
}

func (s *Persistence) makeKey(language, id string) string {
	return fmt.Sprintf("%s%s%s%s", s.basePath, language, PathSeparator, id)
}

func (s *Persistence) mustMarshal(p *models.Policy) []byte {
	b, _ := json.Marshal(p)
	return b
}

func (s *Persistence) unmarshal(id string, kv *store.KVPair) (*models.Policy, error) {
	p := new(models.Policy)
	if err := json.Unmarshal(kv.Value, p); err != nil {
		return s.failure("unmarshal", id, err, true)
	}
	return p, nil
}

func (s *Persistence) failure(op, id string, err error, mustFind bool) (*models.Policy, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to %s policy '%s': %w", op, id, err)
	}
	if mustFind {
		return nil, fmt.Errorf("policy '%s' not found", id)
	}
	return nil, fmt.Errorf("policy '%s' already exists", id)
}

func (s *Persistence) bugFix(in string) string {
	// the Valkeyrie/etcdv3 implementation sometimes removes a leading slash character from the key.
	if _, ok := s.client.(*etcdv3.Store); ok {
		return "/" + in
	}
	return in
}

var (
	readOptions  = &store.ReadOptions{}
	writeOptions = &store.WriteOptions{}
)
