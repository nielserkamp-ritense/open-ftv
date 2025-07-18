package bundles

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

// NewManager instantiates a new bundle manager.
func NewManager(path string, recurse bool, logger *slog.Logger) *Manager {
	m := &Manager{
		path:    path,
		recurse: recurse,
		bundles: make(map[string]*Config),
		logger:  logger,
	}

	m.load()
	return m
}

// Manager represents the interface to manage bundles.
type Manager struct {
	path    string
	recurse bool
	bundles map[string]*Config
	logger  *slog.Logger
}

func (m *Manager) load() {
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

	m.bundles[b.Tag] = &b
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

	m.bundles[b.Tag] = &b
	return nil
}
