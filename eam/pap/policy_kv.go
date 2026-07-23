package pap

// This file is deprecated. Please use the PostgreSQL persistence backend.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// PathSeparator is the standard separator character to use with multi-level keys.
const PathSeparator = "/"

// NewKeyValueDB instantiates a new persistent storage handler for policies with a KV backend.
//
// It creates a CRUD wrapper around the given Valkeyrie Store interface.
// This means it can be used with various distributed KV backends,
// such as Consul, Etcd, Zookeeper, BoltDB, DynamoDB.
// See https://github.com/kvtools/valkeyrie.
//
// Use basePath to define the key prefix to use for the KV backend.
// All policy identifiers will be prefixed with the value of basePath as the KV backend key.
// A trailing PathSeparator character in basePath is automatically appended if it is missing from the input.
func NewKeyValueDB(client store.Store, basePath string) *KeyValueDB {
	return &KeyValueDB{client: client, basePath: convert.ForceSuffix(basePath, PathSeparator)}
}

// KeyValueDB contains the details to manage persistent storage for policies with a KV backend.
type KeyValueDB struct {
	client   store.Store
	basePath string
	mutex    sync.RWMutex
}

// CreatePolicy adds a policy to the store.
//
// Identity is not tracked for the (deprecated) KV backend.
func (s *KeyValueDB) CreatePolicy(ctx context.Context, _ identity.Principal, p *models.Policy) (*models.Policy, error) {
	key := s.makeKey(p.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, _, err := s.client.AtomicPut(ctx, key, s.mustMarshal(p), nil, writeOptions); err != nil {
		return nil, s.failure("create", p.ID(), err, false)
	}
	return p, nil
}

// ReadPolicy retrieves a policy from the store.
func (s *KeyValueDB) ReadPolicy(ctx context.Context, id string) (*models.Policy, uint64, error) {
	key := s.bugFix(s.makeKey(id))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	kv, err := s.client.Get(ctx, key, readOptions)
	switch {
	case errors.Is(err, store.ErrKeyNotFound):
		return nil, 0, nil
	case err != nil:
		return nil, 0, s.failure("read", id, err, true)
	case kv == nil:
		return nil, 0, nil
	}

	p, err2 := s.unmarshal(id, kv)
	if err2 != nil {
		return nil, 0, s.failure("read", id, err2, true)
	}

	return p, kv.LastIndex, nil
}

// ReadPolicyAudit retrieves the audit-log for the identified policy from the database.
//
// Not supported for key-value stores.
func (s *KeyValueDB) ReadPolicyAudit(context.Context, string) ([]policies.AuditEntry, error) {
	return nil, nil
}

// ReadPolicyDeployments retrieves the deployment-log for the identified policy from the database.
//
// Not supported for key-value stores.
func (s *KeyValueDB) ReadPolicyDeployments(context.Context, string) ([]policies.UsageData, error) {
	return nil, nil
}

// ReadPolicyVersions retrieves the versions for the identified policy from the database.
//
// Not supported for key-value stores.
func (s *KeyValueDB) ReadPolicyVersions(context.Context, string) (policies.PolicyVersions, error) {
	return nil, nil
}

// ReadPolicyVersion retrieves a specific version for the identified policy from the database.
//
// Not supported for key-value stores.
func (s *KeyValueDB) ReadPolicyVersion(context.Context, string, int) (*policies.PolicyVersion, error) {
	return nil, nil
}

// UpdatePolicy replaces a policy in the store.
//
// Identity is not tracked for the (deprecated) KV backend.
func (s *KeyValueDB) UpdatePolicy(ctx context.Context, _ identity.Principal, prev *models.Policy, lastIndex uint64, p *models.Policy) (*models.Policy, error) {
	key := s.makeKey(prev.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: s.mustMarshal(prev), LastIndex: lastIndex}
	if _, _, err := s.client.AtomicPut(ctx, key, s.mustMarshal(p), &kv, writeOptions); err != nil {
		return nil, s.failure("update", prev.ID(), err, true)
	}
	return p, nil
}

// DeletePolicy removes a policy from the store.
//
// Identity is not tracked for the (deprecated) KV backend.
func (s *KeyValueDB) DeletePolicy(ctx context.Context, _ identity.Principal, prev *models.Policy, lastIndex uint64) (*models.Policy, error) {
	key := s.makeKey(prev.ID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: s.mustMarshal(prev), LastIndex: lastIndex}
	if _, err := s.client.AtomicDelete(ctx, key, &kv); err != nil {
		return nil, s.failure("delete", prev.ID(), err, true)
	}

	return prev, nil
}

// ListPolicies returns all policies from the store.
func (s *KeyValueDB) ListPolicies(ctx context.Context, language string) ([]*models.Policy, error) {
	key := s.bugFix(s.basePath)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list, err := s.client.List(ctx, key, readOptions)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read policies: %w", err)
	}

	out := make([]*models.Policy, 0, len(list))
	for _, kv := range list {
		if !strings.Contains(kv.Key, "$$$") {
			p := new(models.Policy)
			if err = json.Unmarshal(kv.Value, p); err != nil {
				return nil, fmt.Errorf("failed to unmarshal policies: %w", err)
			}
			if language == "" || p.Language() == language {
				out = append(out, p)
			}
		}
	}

	return out, nil
}

func (s *KeyValueDB) makeKey(id string) string {
	return fmt.Sprintf("%s%s", s.basePath, id)
}

func (s *KeyValueDB) mustMarshal(p *models.Policy) []byte {
	b, _ := json.Marshal(p)
	return b
}

func (s *KeyValueDB) unmarshal(id string, kv *store.KVPair) (*models.Policy, error) {
	p := new(models.Policy)
	if err := json.Unmarshal(kv.Value, p); err != nil {
		return nil, s.failure("unmarshal", id, err, true)
	}
	return p, nil
}

func (s *KeyValueDB) failure(op, id string, err error, mustFind bool) error {
	if err != nil {
		return fmt.Errorf("failed to %s policy '%s': %w", op, id, err)
	}
	if mustFind {
		return fmt.Errorf("policy '%s' not found", id)
	}
	return fmt.Errorf("policy '%s' already exists", id)
}

func (s *KeyValueDB) bugFix(in string) string {
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
