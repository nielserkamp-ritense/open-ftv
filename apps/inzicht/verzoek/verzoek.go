// Package verzoek models Inzicht requests (verzoeken) and persists them through
// the shared Valkeyrie KV pattern (in-memory by default, Postgres/etcd/consul as
// options), mirroring eam/pap/persist.go.
package verzoek

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/kvtools/valkeyrie/store"
)

// PathSeparator separates key segments in the KV store.
const PathSeparator = "/"

// Status is the lifecycle state of an Inzicht request.
type Status string

// The Inzicht request states.
const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusDenied   Status = "denied"
)

// Verzoek is a verstrekker's request to inspect a set of processing activities
// (verwerkingen) held in the afnemer's ADL. It requires approval on the afnemer
// side before the result can be retrieved.
type Verzoek struct {
	ID          string     `json:"id"`
	Verstrekker string     `json:"verstrekker,omitempty"`
	Vanaf       *time.Time `json:"vanaf,omitempty"`
	Tot         *time.Time `json:"tot,omitempty"`
	Doel        string     `json:"doel,omitempty"`
	Afnemer     string     `json:"afnemer,omitempty"` // afnemer scope of the (to-be-)approved disclosure.
	TraceIDs    []string   `json:"trace_ids,omitempty"`
	Status      Status     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	DecidedAt   *time.Time `json:"decided_at,omitempty"`
	DecidedBy   string     `json:"decided_by,omitempty"`
	Reason      string     `json:"reason,omitempty"`
}

// Store is a persistent store for Inzicht requests over any Valkeyrie backend.
type Store struct {
	ctx      context.Context
	client   store.Store
	basePath string
	mu       sync.RWMutex
}

// NewStore wraps a Valkeyrie store with CRUD for Verzoek values under basePath.
func NewStore(ctx context.Context, client store.Store, basePath string) *Store {
	if basePath == "" {
		basePath = "inzicht"
	}
	if basePath[len(basePath)-1] != '/' {
		basePath += PathSeparator
	}
	return &Store{ctx: ctx, client: client, basePath: basePath}
}

// Create stores a new request; it fails if the id already exists.
func (s *Store) Create(v Verzoek) (Verzoek, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, _, err := s.client.AtomicPut(s.ctx, s.key(v.ID), mustMarshal(v), nil, writeOptions); err != nil {
		return Verzoek{}, fmt.Errorf("verzoek: create %q: %w", v.ID, err)
	}
	return v, nil
}

// Get reads a request and its KV revision (for optimistic updates).
func (s *Store) Get(id string) (Verzoek, uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kv, err := s.client.Get(s.ctx, s.key(id), readOptions)
	if err != nil || kv == nil {
		if errors.Is(err, store.ErrKeyNotFound) || kv == nil {
			return Verzoek{}, 0, ErrNotFound
		}
		return Verzoek{}, 0, fmt.Errorf("verzoek: read %q: %w", id, err)
	}
	var v Verzoek
	if err := json.Unmarshal(kv.Value, &v); err != nil {
		return Verzoek{}, 0, fmt.Errorf("verzoek: unmarshal %q: %w", id, err)
	}
	return v, kv.LastIndex, nil
}

// Update atomically replaces a request at the given revision.
func (s *Store) Update(prev Verzoek, lastIndex uint64, v Verzoek) (Verzoek, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	kv := store.KVPair{Key: s.key(prev.ID), Value: mustMarshal(prev), LastIndex: lastIndex}
	if _, _, err := s.client.AtomicPut(s.ctx, s.key(prev.ID), mustMarshal(v), &kv, writeOptions); err != nil {
		return Verzoek{}, fmt.Errorf("verzoek: update %q: %w", prev.ID, err)
	}
	return v, nil
}

// List returns all requests, newest first.
func (s *Store) List() ([]Verzoek, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, err := s.client.List(s.ctx, s.basePath, readOptions)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("verzoek: list: %w", err)
	}

	out := make([]Verzoek, 0, len(list))
	for _, kv := range list {
		var v Verzoek
		if err := json.Unmarshal(kv.Value, &v); err != nil {
			return nil, fmt.Errorf("verzoek: unmarshal list entry: %w", err)
		}
		out = append(out, v)
	}
	sortByCreatedDesc(out)
	return out, nil
}

func (s *Store) key(id string) string { return s.basePath + id }

// ErrNotFound is returned when a request id does not exist.
var ErrNotFound = errors.New("verzoek: not found")

var (
	readOptions  = &store.ReadOptions{}
	writeOptions = &store.WriteOptions{}
)

func mustMarshal(v Verzoek) []byte {
	b, _ := json.Marshal(v)
	return b
}

func sortByCreatedDesc(vs []Verzoek) {
	for i := 1; i < len(vs); i++ {
		for j := i; j > 0 && vs[j].CreatedAt.After(vs[j-1].CreatedAt); j-- {
			vs[j], vs[j-1] = vs[j-1], vs[j]
		}
	}
}
