// Package pap contains all logic for a functional component acting as the Policy Administration Point.
package pap

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kvtools/valkeyrie/store"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// ErrNoPersistence is returned by New when no persistence backend was configured via
// WithPgPool(), WithKeyValueDB(), or WithPolicyDB().
var ErrNoPersistence = errors.New("pap: no persistence backend configured")

// logArgsCapacity is the pre-allocated size of the slog argument slice built in logInitialized.
const logArgsCapacity = 8

// PAP represents the interface for caching and retrieving policies.
type PAP struct {
	languageType  models.Language
	recurse       bool
	policyStore   string
	language      string
	ctx           context.Context
	logger        *slog.Logger
	languageDB    LanguagePersister
	tagDB         TagPersister
	policyDB      PolicyPersister
	policyWatcher *fsnotify.Watcher
	policyTimer   *time.Timer
	policyUpdates map[string]struct{}
	policyDeletes map[string]struct{}
	bundleDB      bundles.Persister
	kvStore       store.Store
	migrateSource string
	migrateDB     string
	migrateAuto   bool
	migrateSteps  int
	eventSinks    []models.EventSink
	eventMutex    sync.RWMutex
	deployMutex   sync.RWMutex
}

// New instantiates a new policy cache.
//
// The optional context can be used to signal an orderly shutdown.
//
// A PAP requires a persistence backend: use WithPgPool(), WithKeyValueDB(), or WithPolicyDB() to connect one.
// Without one, New returns ErrNoPersistence.
// WithFileStore() additionally loads policies from local files into whichever backend is configured.
// WithMigration() configures a database migration but does not run it — New only wires dependencies;
// call Migrate() explicitly once New succeeds.
func New(ctx context.Context, logger *slog.Logger, options ...Option) (_ *PAP, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	w := newFileWatcher()

	p := &PAP{
		ctx:           ctx,
		logger:        logger,
		policyWatcher: w,
		policyUpdates: make(map[string]struct{}),
		policyDeletes: make(map[string]struct{}),
		eventSinks:    make([]models.EventSink, 0),
	}

	// watchFiles() takes ownership of w and closes it once started; if we fail before that, closing it here.
	defer func() {
		if err != nil && w != nil {
			_ = w.Close()
		}
	}()

	for i := range options {
		options[i](p)
	}

	if p.policyDB == nil {
		logger.Error("pap: no persistence backend configured")
		return nil, ErrNoPersistence
	}

	haveStore := p.policyStore != "" && p.policyStore != "/"
	if haveStore && w != nil {
		// have file storage and watcher: watch file changes.
		go p.watchFiles()
	}

	p.logInitialized(haveStore)

	return p, nil
}

// newFileWatcher creates a filesystem watcher, or returns nil if file handles are exhausted.
func newFileWatcher() *fsnotify.Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil
	}

	return w
}

func (p *PAP) logInitialized(haveStore bool) {
	if !p.logger.Enabled(p.ctx, slog.LevelInfo) {
		return
	}

	args := make([]any, 0, logArgsCapacity)
	if haveStore {
		args = append(args, "policyStore", p.policyStore, "recurse", p.recurse)
	}

	if p.kvStore == nil {
		args = append(args, "persistence", true)
	}

	p.logger.Info("pap initialized", args...)
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
