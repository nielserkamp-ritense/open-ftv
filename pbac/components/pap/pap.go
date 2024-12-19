// Package pap contains all logic for a functional component acting as the Policy Administration Point.
package pap

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// PAP represents the interface for caching and retrieving policies.
type PAP interface {
	Add(in Policy) (Policy, error)
	Replace(in Policy) (Policy, error)
	Remove(id string) (Policy, error)
	Get(id string) (Policy, error)
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
		policies: make(map[string]Policy),
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
// An error is returned if the policy key already exists.
func (p *pap) Add(in Policy) (out Policy, err error) {
	p.mutex.Lock()
	if _, ok := p.policies[in.ID()]; ok {
		err = fmt.Errorf("cache policy '%s' already exists", in.ID())
	} else {
		out = in
		p.policies[out.ID()] = out
	}
	p.mutex.Unlock()

	if out != nil && p.events != nil {
		p.events.Handle(models.PolicyAdded, out.ID())
	}
	return
}

// Replace modifies a policy in the cache with a newer version.
//
// An error is returned if the policy key doesn't exist.
func (p *pap) Replace(in Policy) (out Policy, err error) {
	p.mutex.Lock()
	if _, ok := p.policies[in.ID()]; !ok {
		err = fmt.Errorf("cache policy '%s' not found", in.ID())
	} else {
		out = in
		p.policies[out.ID()] = out
	}
	p.mutex.Unlock()

	if out != nil && p.events != nil {
		p.events.Handle(models.PolicyReplaced, out.ID())
	}
	return
}

// Remove removes a policy from the cache.
//
// An error is returned if the policy key doesn't exist.
func (p *pap) Remove(id string) (out Policy, err error) {
	p.mutex.Lock()
	if old, ok := p.policies[id]; !ok {
		err = fmt.Errorf("cache policy '%s' not found", id)
	} else {
		out = old
		delete(p.policies, id)
	}
	p.mutex.Unlock()

	if out != nil && p.events != nil {
		p.events.Handle(models.PolicyRemoved, out.ID())
	}
	return
}

// Get retrieves a policy from the cache, or an error if the policy key doesn't exist.
func (p *pap) Get(id string) (Policy, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if data, ok := p.policies[id]; ok {
		return data, nil
	}

	return nil, fmt.Errorf("cache policy '%s' not found", id)
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
	policies map[string]Policy
	updates  []string
	deletes  []string
	events   models.EventSink
	mutex    sync.RWMutex
}
