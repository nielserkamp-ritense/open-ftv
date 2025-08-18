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
	languageType  models.Language
	recurse       bool
	policyStore   string
	language      string
	ctx           context.Context
	logger        *slog.Logger
	watcher       *fsnotify.Watcher
	wTimer        *time.Timer
	updates       map[string]struct{}
	deletes       map[string]struct{}
	eventSinks    []models.EventSink
	kvStore       store.Store
	kvDB          *KeyValueDB
	postgresDB    *PostgresDB
	migrateSource string
	migrateDB     string
	migrateAuto   bool
	migrateSteps  int
	deployer      *bundles.Persistence
	eventMutex    sync.RWMutex
	deployMutex   sync.RWMutex
}

// New instantiates a new policy cache.
//
// The optional context can be used to signal an orderly shutdown.
//
// By default, a PAP uses an in-memory key-value cache.
// Use the WithKeyValueDB() or WithPostgresDB() option to connect a PAP to persistent storage.
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

	haveStore := p.policyStore != "" && p.policyStore != "/"
	persist := p.kvStore != nil || p.postgresDB != nil

	if persist && p.migrateSource != "" && p.migrateDB != "" && (p.migrateAuto || p.migrateSteps != 0) {
		// persistence and migration configured: run the migration.
		if err = p.migration(); err != nil {
			return nil
		}
	}

	if !persist {
		// no persistence configured: force in-memory storage.
		WithKeyValueDB(memory.New(), "")(p)
	}

	if haveStore && w != nil {
		// have file storage and watcher: watch file changes.
		go p.watchFiles()
	}

	if p.logger.Enabled(nil, slog.LevelInfo) {
		args := make([]any, 0, 8)
		if haveStore {
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

// AddEventSink adds an event processor to the PAP.
func (p *PAP) AddEventSink(events models.EventSink) {
	p.eventMutex.Lock()
	p.eventSinks = append(p.eventSinks, events)
	p.eventMutex.Unlock()
}

func (p *PAP) sendEvent(eventType models.EventType, key string) {
	p.eventMutex.Lock()
	for i := range p.eventSinks {
		p.eventSinks[i].Handle(eventType, key)
	}
	defer p.eventMutex.Unlock()
}
