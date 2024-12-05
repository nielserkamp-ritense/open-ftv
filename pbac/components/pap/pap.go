// Package pap contains all logic for a functional component acting as the Policy Administration Point.
package pap

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// PAP represents the interface for caching and retrieving policies.
type PAP interface {
	Add(key string, reader io.Reader) error
	Replace(key string, reader io.Reader) error
	Remove(key string) error
	Get(key string) (io.Reader, error)
	ListAllKeys() []string
	LoadFromStore(path string, recurse bool)
}

// New instantiates a new policy cache.
//
// The optional context can be used to signal app shutdown by closing it,
// so the PAP can clean up long-running go-routines and other resources.
func New(ctx context.Context, logger *slog.Logger, events models.EventSink) PAP {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		w = nil // this means file handles are exhausted!
	}

	if ctx == nil {
		ctx = context.Background()
	}

	c := &pap{
		ctx:      ctx,
		logger:   logger,
		events:   events,
		policies: make(map[string][]byte),
		watcher:  w,
	}

	if w != nil {
		go c.watchFiles()
	}

	c.logger.Info("pap initialized")
	return c
}

// Add adds a policy to the cache.
//
// If the input reader is nil, the function returns duccessfully without doing anything.
//
// An error is returned if the policy key already exists.
func (p *pap) Add(key string, reader io.Reader) error {
	if reader == nil {
		return nil
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	p.mutex.Lock()
	if _, ok := p.policies[key]; ok {
		err = fmt.Errorf("cache policy '%s' already exists", key)
	} else {
		p.policies[key] = data
	}
	p.mutex.Unlock()

	if err == nil && p.events != nil {
		p.events.Handle(models.PolicyAdded, key)
	}
	return err
}

// Replace modifies a policy in the cache with a newer version.
//
// If the input reader is nil, the function returns successfully without doing anything.
//
// An error is returned if the policy key doesn't exist.
func (p *pap) Replace(key string, reader io.Reader) error {
	if reader == nil {
		return nil
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	p.mutex.Lock()
	if _, ok := p.policies[key]; !ok {
		err = fmt.Errorf("cache policy '%s' not found", key)
	} else {
		p.policies[key] = data
	}
	p.mutex.Unlock()

	if err == nil && p.events != nil {
		p.events.Handle(models.PolicyReplaced, key)
	}
	return err
}

// Remove removes a policy from the cache.
//
// An error is returned if the policy key doesn't exist.
func (p *pap) Remove(key string) error {
	var err error

	p.mutex.Lock()
	if _, ok := p.policies[key]; !ok {
		err = fmt.Errorf("cache policy '%s' not found", key)
	} else {
		delete(p.policies, key)
	}
	p.mutex.Unlock()

	if err == nil && p.events != nil {
		p.events.Handle(models.PolicyRemoved, key)
	}
	return err
}

// Get retrieves a policy from the cache, or an error if the policy key doesn't exist.
func (p *pap) Get(key string) (io.Reader, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if data, ok := p.policies[key]; ok {
		// we return a reader on a deep copy of the data, so changes in the cache do not affect it.
		s := string(data)
		return bytes.NewBufferString(s), nil
	}

	return nil, fmt.Errorf("cache policy '%s' not found", key)
}

// ListAllKeys returns a list of all cached policy keys.
func (p *pap) ListAllKeys() []string {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	out := make([]string, 0, len(p.policies))
	for k := range p.policies {
		out = append(out, k)
	}

	slices.Sort(out)
	return out
}

type pap struct {
	recurse  bool
	path     string
	ctx      context.Context
	logger   *slog.Logger
	watcher  *fsnotify.Watcher
	wTimer   *time.Timer
	policies map[string][]byte
	updates  []string
	deletes  []string
	events   models.EventSink
	mutex    sync.RWMutex
}
