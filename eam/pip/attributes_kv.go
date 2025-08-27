package pip

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// NewAttributeStore instantiates a new persistent storage handler for attributes.
//
// It creates a CRUD wrapper around the given Valkeyrie Store interface.
// This means it can be used with various distributed KV backends,
// such as Consul, Etcd, Zookeeper, BoltDB, DynamoDB.
// See https://github.com/kvtools/valkeyrie.
//
// Use basePath to define the key prefix to use for the backend KV store.
// All attribute keys will be prefixed with the value of basePath as the KV store's key.
// A trailing pathSeparator character in basePath is automatically appended if it is missing from the input.
//
// The given context is passed in every call to the KV backend.
func NewAttributeStore(client store.Store, basePath string) AttributePersister {
	return &attributeStore{kvWrapper{client: client, basePath: convert.ForceSuffix(basePath, PathSeparator)}}
}

// CreateAttribute implements the AttributePersister interface.
func (s *attributeStore) CreateAttribute(ctx context.Context, a *models.Attribute) (*models.Attribute, error) {
	key := s.makeKey(a.Key())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, _, err := s.client.AtomicPut(ctx, key, marshalAttribute(a), nil, writeOptions); err != nil {
		return nil, s.failure("create", a.Key(), err, false)
	}
	return a, nil
}

// ReadAttribute implements the AttributePersister interface.
func (s *attributeStore) ReadAttribute(ctx context.Context, id string) (*models.Attribute, uint64, error) {
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

	a, err2 := unmarshalAttribute(kv.Value)
	if err2 != nil {
		return nil, 0, s.failure("read", id, err2, true)
	}

	return a, kv.LastIndex, nil
}

// UpdateAttribute implements the AttributePersister interface.
func (s *attributeStore) UpdateAttribute(ctx context.Context, prev *models.Attribute, lastIndex uint64, a *models.Attribute) (*models.Attribute, error) {
	key := s.makeKey(prev.Key())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: marshalAttribute(prev), LastIndex: lastIndex}
	if _, _, err := s.client.AtomicPut(ctx, key, marshalAttribute(a), &kv, writeOptions); err != nil {
		return nil, s.failure("update", prev.Key(), err, true)
	}
	return a, nil
}

// DeleteAttribute implements the AttributePersister interface.
func (s *attributeStore) DeleteAttribute(ctx context.Context, prev *models.Attribute, lastIndex uint64) (*models.Attribute, error) {
	key := s.makeKey(prev.Key())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	kv := store.KVPair{Key: key, Value: marshalAttribute(prev), LastIndex: lastIndex}
	if _, err := s.client.AtomicDelete(ctx, key, &kv); err != nil {
		return nil, s.failure("delete", prev.Key(), err, true)
	}

	return prev, nil
}

// ListAttributes implements the AttributePersister interface.
func (s *attributeStore) ListAttributes(ctx context.Context) ([]*models.Attribute, error) {
	key := s.bugFix(s.basePath)

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	list, err := s.client.List(ctx, key, readOptions)
	if err != nil || len(list) == 0 {
		if errors.Is(err, store.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read attributes: %w", err)
	}

	out := make([]*models.Attribute, 0, len(list))
	for _, kv := range list {
		a, err2 := unmarshalAttribute(kv.Value)
		if err2 != nil {
			return nil, fmt.Errorf("failed to unmarshal attributes: %w", err2)
		}
		out = append(out, a)
	}

	return out, nil
}

func marshalAttribute(a *models.Attribute) []byte {
	b, _ := json.Marshal(toAttribute(a))
	return b
}

func toAttribute(a *models.Attribute) *attribute {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)

	q := a.Value()
	if err := enc.Encode(&q); err != nil {
		panic(err)
	}
	v := base64.StdEncoding.EncodeToString(buf.Bytes())

	buf.Reset()

	q = a.Original()
	if err := enc.Encode(&q); err != nil {
		panic(err)
	}
	o := base64.StdEncoding.EncodeToString(buf.Bytes())

	if o == v {
		o = ""
	}

	return &attribute{
		Key:         a.Key(),
		Value:       v,
		Original:    o,
		Type:        a.Type(),
		Title:       a.Title(),
		Description: a.Description(),
		Tags:        a.Tags(),
	}
}

func unmarshalAttribute(data []byte) (*models.Attribute, error) {
	a := &attribute{}
	if err := json.Unmarshal(data, a); err != nil {
		return nil, err
	}
	return fromAttribute(a)
}

func fromAttribute(a *attribute) (*models.Attribute, error) {
	buf := &bytes.Buffer{}
	dec := gob.NewDecoder(buf)

	d, err := base64.StdEncoding.DecodeString(a.Value)
	if err != nil {
		return nil, err
	}
	buf.Write(d)

	var v, o any
	if err = dec.Decode(&v); err != nil {
		return nil, err
	}

	if a.Original == "" {
		o = v
	} else {
		d, err = base64.StdEncoding.DecodeString(a.Original)
		if err != nil {
			return nil, err
		}

		buf.Reset()
		buf.Write(d)
		if err = dec.Decode(&o); err != nil {
			return nil, err
		}
	}

	return models.NewOriginalAttribute(a.Key, v, o, a.Type).WithTitle(a.Title).WithDescription(a.Description).WithTags(a.Tags...), nil
}

func (s *attributeStore) failure(op, id string, err error, mustFind bool) error {
	if err != nil {
		return fmt.Errorf("failed to %s attribute '%s': %w", op, id, err)
	}
	if mustFind {
		return fmt.Errorf("attribute '%s' not found", id)
	}
	return fmt.Errorf("attribute '%s' already exists", id)
}

type attributeStore struct {
	kvWrapper
}

type attribute struct {
	Key         string   `json:"key"`
	Value       string   `json:"value"`
	Original    string   `json:"original,omitempty"`
	Type        string   `json:"type,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}
