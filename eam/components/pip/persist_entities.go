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

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// EntityPersistence represents the interface to manage persistent storage for attributes.
type EntityPersistence interface {
	Create(models.Entity) (models.Entity, error)
	Read(string) (models.Entity, uint64, error)
	Update(models.Entity, uint64, models.Entity) (models.Entity, error)
	Delete(models.Entity, uint64) (models.Entity, error)
	List() ([]models.Entity, error)
}

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
func NewEntityStore(ctx context.Context, client store.Store, basePath string) EntityPersistence {
	if basePath != "" && !strings.HasSuffix(basePath, pathSeparator) {
		basePath += pathSeparator
	}
	return &entityStore{wrapper{ctx: ctx, client: client, basePath: basePath}}
}

// Create implements the EntityPersistence interface.
func (s *entityStore) Create(e models.Entity) (models.Entity, error) {
	key := s.makeKey(e.UID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, _, err := s.client.AtomicPut(s.ctx, key, marshalEntity(e), nil, writeOptions); err != nil {
		return nil, s.failure("create", e.UID(), err, false)
	}
	return e, nil
}

// Read implements the EntityPersistence interface.
func (s *entityStore) Read(id string) (models.Entity, uint64, error) {
	key := s.bugFix(s.makeKey(id))

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	kv, err := s.client.Get(s.ctx, key, readOptions)
	if err != nil || kv == nil {
		return nil, 0, s.failure("read", id, err, true)
	}

	e, err2 := unmarshalEntity(kv.Value)
	if err2 != nil {
		return nil, 0, s.failure("read", id, err2, true)
	}

	return e, kv.LastIndex, nil
}

// Update implements the EntityPersistence interface.
func (s *entityStore) Update(prev models.Entity, lastIndex uint64, e models.Entity) (models.Entity, error) {
	key := s.makeKey(prev.UID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: marshalEntity(prev), LastIndex: lastIndex}
	if _, _, err := s.client.AtomicPut(s.ctx, key, marshalEntity(e), &kv, writeOptions); err != nil {
		return nil, s.failure("update", prev.UID(), err, true)
	}
	return e, nil
}

// Delete implements the EntityPersistence interface.
func (s *entityStore) Delete(prev models.Entity, lastIndex uint64) (models.Entity, error) {
	key := s.makeKey(prev.UID())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: marshalEntity(prev), LastIndex: lastIndex}
	if _, err := s.client.AtomicDelete(s.ctx, key, &kv); err != nil {
		return nil, s.failure("delete", prev.UID(), err, true)
	}

	return prev, nil
}

// List implements the EntityPersistence interface.
func (s *entityStore) List() ([]models.Entity, error) {
	key := s.bugFix(s.basePath)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list, err := s.client.List(s.ctx, key, readOptions)
	if err != nil || len(list) == 0 {
		if errors.Is(err, store.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read entities: %w", err)
	}

	out := make([]models.Entity, 0, len(list))
	for _, kv := range list {
		e, err2 := unmarshalEntity(kv.Value)
		if err2 != nil {
			return nil, fmt.Errorf("failed to unmarshalEntity entities: %w", err2)
		}
		out = append(out, e)
	}

	return out, nil
}

func marshalEntity(e models.Entity) []byte {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	var v string

	if s := e.Attributes(); s != nil {
		q := make([]*attribute, 0)
		e.Attributes().IterateAttributes(func(attr models.Attribute) {
			q = append(q, toAttribute(attr))
		})

		slices.SortFunc(q, func(a, b *attribute) int {
			return strings.Compare(a.Key, b.Key)
		})

		if err := enc.Encode(&q); err != nil {
			panic(err)
		}
		v = base64.StdEncoding.EncodeToString(buf.Bytes())
	}

	b, _ := json.Marshal(&entity{Type: e.Type(), ID: e.ID(), Attributes: v, Parents: e.Parents()})
	return b
}

func unmarshalEntity(data []byte) (models.Entity, error) {
	buf := &bytes.Buffer{}
	dec := gob.NewDecoder(buf)

	e := &entity{}
	if err := json.Unmarshal(data, e); err != nil {
		return nil, err
	}

	d, err := base64.StdEncoding.DecodeString(e.Attributes)
	if err != nil {
		return nil, err
	}
	buf.Write(d)

	m := make([]*attribute, 0)
	if err = dec.Decode(&m); err != nil {
		return nil, err
	}

	s := models.NewAttributeSet()
	for i := range m {
		a, err2 := fromAttribute(m[i])
		if err2 != nil {
			return nil, err2
		}
		s.AddOriginalAttribute(a.Key(), a.Value(), a.Original(), a.Type())
	}

	return models.NewEntity(e.Type, e.ID, s, e.Parents...), nil
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
	wrapper
}

type entity struct {
	Type       string   `json:"type"`
	ID         string   `json:"id"`
	Attributes string   `json:"attributes,omitempty"`
	Parents    []string `json:"parents,omitempty"`
}
