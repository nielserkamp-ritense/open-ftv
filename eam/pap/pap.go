// Package pap contains all logic for a functional component acting as the Policy Administration Point.
package pap

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// PAP represents the interface for caching and retrieving policies.
type PAP struct {
	languageType models.Language
	recurse      bool
	policyStore  string
	language     string
	ctx          context.Context
	logger       *slog.Logger
	watcher      *fsnotify.Watcher
	wTimer       *time.Timer
	updates      map[string]struct{}
	deletes      map[string]struct{}
	eventSinks   []models.EventSink
	store        store.Store
	persist      *Persistence
	deployer     *bundles.Persistence
	mutex        sync.RWMutex
}

// New instantiates a new policy cache.
//
// The optional context can be used to signal an orderly shutdown.
//
// By default, a PAP uses an in-memory key-value cache.
// Use the WithPersistence() option to connect a PAP to persistent storage.
func New(ctx context.Context, logger *slog.Logger, options ...Option) *PAP {
	if ctx == nil {
		ctx = context.Background()
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		w = nil // this means file handles are exhausted!
	}

	p := &PAP{
		ctx:        ctx,
		logger:     logger,
		watcher:    w,
		updates:    make(map[string]struct{}),
		deletes:    make(map[string]struct{}),
		eventSinks: make([]models.EventSink, 0),
	}

	for i := range options {
		options[i](p)
	}

	persist := p.store != nil
	if !persist {
		// force in-memory storage.
		WithPersistence(memory.New(), "")(p)
	}

	if w != nil {
		go p.watchFiles()
	}

	if p.logger.Enabled(nil, slog.LevelInfo) {
		args := make([]any, 0, 8)
		if p.policyStore != "" {
			args = append(args, "policyStore", p.policyStore, "recurse", p.recurse)
		}
		if persist {
			args = append(args, "persistence", true)
		}
		p.logger.Info("pap initialized", args...)
	}

	return p
}

// Language returns the default policy language for the PAP.
func (p *PAP) Language() models.Language {
	return p.languageType
}

// Create adds a policy to cache/storage.
//
// An error is returned if the policy-id already exists.
func (p *PAP) Create(in *models.Policy) (out *models.Policy, err error) {
	p.mutex.Lock()
	out, err = p.persist.Create(in)
	p.mutex.Unlock()

	if err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.PolicyAdded, out.Key())
	}
	return
}

// Read retrieves a policy from cache/storage.
//
// An error is returned if the policy-id doesn't exist.
func (p *PAP) Read(language, id string) (*models.Policy, uint64, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.persist.Read(language, id)
}

// Update modifies a policy in cache/storage with a newer version.
//
// An error is returned if the policy-id doesn't exist.
func (p *PAP) Update(prev *models.Policy, lastIndex uint64, in *models.Policy) (out *models.Policy, err error) {
	p.mutex.Lock()
	out, err = p.persist.Update(prev, lastIndex, in)
	p.mutex.Unlock()

	if err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.PolicyReplaced, out.Key())
	}

	return
}

// Delete removes a policy from cache/storage.
//
// An error is returned if the policy key doesn't exist.
func (p *PAP) Delete(prev *models.Policy, lastIndex uint64) (out *models.Policy, err error) {
	p.mutex.Lock()
	out, err = p.persist.Delete(prev, lastIndex)
	p.mutex.Unlock()

	if err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.PolicyRemoved, out.Key())
	}

	return
}

// List returns a sorted list of all cached/stored policy keys.
//
// If the optional language parameter is supplied,
// the function lists all policies with that language.
// Otherwise, all policies, regardless of language, will be listed.
func (p *PAP) List(language string) ([]*models.Policy, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.persist.List(language)
}

// NewDeployment creates a new deployment in the store.
func (p *PAP) NewDeployment(description string, manager *bundles.Manager) (*bundles.Deployment, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	d, err := p.deployer.Generate(description)
	if err != nil {
		return nil, err
	}

	manager.Run(d, p.deployer)
	return d, nil
}

// RestartDeployment checks if a bundle deployment was interrupted and restarts the run if so.
func (p *PAP) RestartDeployment(manager *bundles.Manager) {
	if p.deployer != nil {
		if d, err2 := p.deployer.LastDeployment(); err2 == nil {
			if s := d.Status(); s != bundles.Failed && s != bundles.Completed {
				manager.Run(d, p.deployer)
			}
		}
	}
}

// LastDeployment retrieves the last deployment from the store.
func (p *PAP) LastDeployment() (*bundles.Deployment, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.deployer.LastDeployment()
}

// ReadDeployment retrieves a deployment from the store.
func (p *PAP) ReadDeployment(version uint64) (*bundles.Deployment, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.deployer.ReadDeployment(version)
}

// ListDeployments retrieves all deployments from the store.
func (p *PAP) ListDeployments() ([]*bundles.Deployment, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.deployer.ListDeployments()
}

// AddEventSink adds an event processor to the PAP.
func (p *PAP) AddEventSink(events models.EventSink) {
	p.mutex.Lock()
	p.eventSinks = append(p.eventSinks, events)
	p.mutex.Unlock()
}

func (p *PAP) sendEvent(eventType models.EventType, key string) {
	p.mutex.RLock()
	for i := range p.eventSinks {
		p.eventSinks[i].Handle(eventType, key)
	}
	p.mutex.RUnlock()
}
