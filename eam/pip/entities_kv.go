package pip

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/goccy/go-json"
	"github.com/kvtools/etcdv3"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// NewEntityStore instantiates a new persistent storage handler for attributes.
//
// It creates a CRUD entity store around the given Valkeyrie Store interface.
// This means it can be used with various distributed KV backends,
// such as Consul, Etcd, Zookeeper, BoltDB, DynamoDB.
// See https://github.com/kvtools/valkeyrie.
//
// Use basePath to define the key prefix to use for the backend KV store.
// All attribute keys will be prefixed with the value of basePath as the KV store's key.
// A trailing pathSeparator character in basePath is automatically appended if it is missing from the input.
//
// The given context is passed in every call to the KV backend.
func NewEntityStore(client store.Store, basePath string) EntityPersister {
	return &entityStore{kvWrapper{client: client, basePath: convert.ForceSuffix(basePath, PathSeparator)}}
}

// CreateEntity implements the EntityPersister interface.
func (s *entityStore) CreateEntity(ctx context.Context, e *models.Entity) (*models.Entity, error) {
	key := s.makeKey(e.UID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, _, err := s.client.AtomicPut(ctx, key, marshalEntity(e), nil, writeOptions); err != nil {
		return nil, s.failure("create", e.UID(), err, false)
	}
	return e, nil
}

// ReadEntity implements the EntityPersister interface.
func (s *entityStore) ReadEntity(ctx context.Context, ns, id string) (*models.Entity, uint64, error) {
	key := s.bugFix(s.makeKey(models.EntityUID(ns, id)))

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

	e, err2 := unmarshalEntity(kv.Value)
	if err2 != nil {
		return nil, 0, s.failure("read", id, err2, true)
	}

	return e, kv.LastIndex, nil
}

// ReadEntityAudit retrieves the audit-log for the identified entity from the database.
//
// Not implemented in key-value stores.
func (s *entityStore) ReadEntityAudit(context.Context, string, string) ([]oas.AuditEntry, error) {
	return nil, nil
}

// ReadEntityDeployments retrieves the deployment-log for the identified entity from the database.
//
// Not implemented in key-value stores.
func (s *entityStore) ReadEntityDeployments(context.Context, string, string) ([]oas.UsageData, error) {
	return nil, nil
}

// ReadEntityVersions retrieves the versions for the identified entity from the database.
//
// Not supported for key-value stores.
func (s *entityStore) ReadEntityVersions(context.Context, string, string) (oas.EntityVersions, error) {
	return nil, nil
}

// ReadEntityVersion retrieves a specific version for the identified entity from the database.
//
// Not supported for key-value stores.
func (s *entityStore) ReadEntityVersion(context.Context, string, string, int) (*oas.EntityVersion, error) {
	return nil, nil
}

// UpdateEntity implements the EntityPersister interface.
func (s *entityStore) UpdateEntity(ctx context.Context, prev *models.Entity, lastIndex uint64, e *models.Entity) (*models.Entity, error) {
	key := s.makeKey(prev.UID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: marshalEntity(prev), LastIndex: lastIndex}
	if _, _, err := s.client.AtomicPut(ctx, key, marshalEntity(e), &kv, writeOptions); err != nil {
		return nil, s.failure("update", prev.UID(), err, true)
	}
	return e, nil
}

// DeleteEntity implements the EntityPersister interface.
func (s *entityStore) DeleteEntity(ctx context.Context, prev *models.Entity, lastIndex uint64) (*models.Entity, error) {
	key := s.makeKey(prev.UID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: marshalEntity(prev), LastIndex: lastIndex}
	if _, err := s.client.AtomicDelete(ctx, key, &kv); err != nil {
		return nil, s.failure("delete", prev.UID(), err, true)
	}

	return prev, nil
}

// ListEntities implements the EntityPersister interface.
func (s *entityStore) ListEntities(ctx context.Context) ([]*models.Entity, error) {
	key := s.bugFix(s.basePath)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list, err := s.client.List(ctx, key, readOptions)
	if err != nil || len(list) == 0 {
		if errors.Is(err, store.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read entities: %w", err)
	}

	out := make([]*models.Entity, 0, len(list))
	for _, kv := range list {
		e, err2 := unmarshalEntity(kv.Value)
		if err2 != nil {
			return nil, fmt.Errorf("failed to unmarshalEntity entities: %w", err2)
		}
		out = append(out, e)
	}

	return out, nil
}

func marshalEntity(e *models.Entity) []byte {
	var v string
	if s := e.Attributes(); s != nil {
		v = base64.StdEncoding.EncodeToString(gobEncodeAttributes(s))
	}

	b, _ := json.Marshal(&entity{Type: e.Type(), ID: e.ID(), Attributes: v, Parents: e.Parents()})
	return b
}

// use gob encode so we don't lose the actual value and original value in lossy JSON encoding!
func gobEncodeAttributes(s *models.AttributeSet) []byte {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	q := make([]*attribute, 0)
	s.IterateAttributes(func(attr *models.Attribute) {
		q = append(q, toAttribute(attr))
	})

	if len(q) == 0 {
		return nil
	}

	slices.SortFunc(q, func(a, b *attribute) int {
		return strings.Compare(a.Key, b.Key)
	})

	if err := enc.Encode(&q); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func unmarshalEntity(data []byte) (*models.Entity, error) {
	e := &entity{}
	if err := json.Unmarshal(data, e); err != nil {
		return nil, err
	}

	b, err := base64.StdEncoding.DecodeString(e.Attributes)
	if err != nil {
		return nil, err
	}

	s, err2 := gobDecodeAttributes(b)
	if err2 != nil {
		return nil, err2
	}

	return models.NewEntity(e.Type, e.ID, s, e.Parents...), nil
}

// use gob decode so we can restore the value and original value without loss.
func gobDecodeAttributes(b []byte) (*models.AttributeSet, error) {
	if len(b) == 0 {
		return models.NewAttributeSet(), nil
	}

	buf := &bytes.Buffer{}
	dec := gob.NewDecoder(buf)

	buf.Write(b)

	m := make([]*attribute, 0)
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}

	s := models.NewAttributeSet()
	for i := range m {
		a, err2 := fromAttribute(m[i])
		if err2 != nil {
			return nil, err2
		}
		_, _ = s.AddAttribute(a)
	}

	return s, nil
}

func (s *entityStore) failure(op, id string, err error, mustFind bool) error {
	if err != nil {
		return fmt.Errorf("failed to %s entity '%s': %w", op, id, err)
	}
	if mustFind {
		return fmt.Errorf("entity '%s' not found", id)
	}
	return fmt.Errorf("entity '%s' already exists", id)
}

func (s *entityStore) bugFix(in string) string {
	// the Valkeyrie/etcdv3 implementation sometimes removes a leading slash character from the key.
	if _, ok := s.client.(*etcdv3.Store); ok {
		return "/" + in
	}
	return in
}

type entityStore struct {
	kvWrapper
}

type entity struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Attributes string   `json:"attributes,omitempty"`
	Parents    []string `json:"parents,omitempty"`
}
