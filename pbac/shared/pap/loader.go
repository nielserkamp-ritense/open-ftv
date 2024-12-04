package pap

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LoadFromStore loads all policies from the store.
func (p *pap) LoadFromStore(path string, recurse bool) {
	p.path = path
	p.recurse = recurse
	p.clearWatcher()

	if p.path == "" {
		return
	}

	if err := filepath.WalkDir(path, p.loadPolicy); err != nil {
		p.logger.Error("pap: error loading policies", "path", path, "err", err)
	}
}

func (p *pap) loadPolicy(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.IsDir() {
		if p.recurse || path == p.path {
			if p.watcher != nil {
				_ = p.watcher.Add(path)
			}
			return nil
		}
		return filepath.SkipDir
	}

	if strings.HasSuffix(path, ".gitkeep") {
		return nil
	}

	f, err2 := os.Open(path)
	if err2 != nil {
		return err2
	}

	defer f.Close()
	return p.Add(d.Name(), f)
}
