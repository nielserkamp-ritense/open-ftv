package bundles

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

// NewManager instantiates a new bundle manager.
func NewManager(ctx context.Context, logger *slog.Logger, opts ...Option) *Manager {
	m := &Manager{
		ctx:           ctx,
		logger:        logger,
		bundles:       make(map[string]*Config),
		workers:       runtime.NumCPU(),
		bundleTimeout: time.Minute,
	}

	for i := range opts {
		opts[i](m)
	}

	m.client = &http.Client{Timeout: m.bundleTimeout}

	m.load()
	return m
}

// Bundles returns a list of configured bundles, ordered by unique identifier.
func (m *Manager) Bundles() []*Config {
	out := make([]*Config, 0, len(m.bundles))
	for _, v := range m.bundles {
		out = append(out, v)
	}

	slices.SortFunc(out, func(a, b *Config) int {
		return strings.Compare(a.ID, b.ID)
	})
	return out
}

// Manager contains the data and logic to manage bundles.
type Manager struct {
	ctx           context.Context
	logger        *slog.Logger
	path          string
	recurse       bool
	workers       int
	stageDelay    time.Duration
	bundleTimeout time.Duration
	policies      PolicyLister
	attributes    AttributeLister
	entities      EntityLister
	relations     RelationLister
	bundles       map[string]*Config
	client        *http.Client
}

func (m *Manager) load() {
	if m.path == "" {
		return
	}

	if err := filepath.WalkDir(m.path, m.processEntry); err != nil {
		m.logger.Error("bundle-manager: error loading bundle configurations", "path", m.path, "err", err)
	}
}

func (m *Manager) processEntry(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.IsDir() {
		if m.recurse || path == m.path {
			return nil
		}
		return filepath.SkipDir
	}

	switch io.SnifFile(path) {
	case io.MimeTypeYAML:
		return m.loadYAML(path)
	case io.MimeTypeJSON:
		return m.loadJSON(path)
	default:
		return nil
	}
}

func (m *Manager) loadYAML(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var b Config
	if err = yaml.NewDecoder(f).Decode(&b); err != nil {
		return err
	}

	m.bundles[b.ID] = &b
	return nil
}

func (m *Manager) loadJSON(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var b Config
	if err = json.NewDecoder(f).Decode(&b); err != nil {
		return err
	}

	m.bundles[b.ID] = &b
	return nil
}
