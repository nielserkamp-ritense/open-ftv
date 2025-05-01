package pap

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// LoadFiles loads all policies from the local file store.
func (p *pap) LoadFiles() {
	if p.policyStore == "" {
		return
	}

	p.clearWatcher()

	if err := filepath.WalkDir(p.policyStore, p.loadPolicy); err != nil {
		p.logger.Error("pap: error loading policies", "policyStore", p.policyStore, "err", err)
	}
}

func (p *pap) loadPolicy(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.IsDir() {
		if p.recurse || path == p.policyStore {
			if p.watcher != nil {
				_ = p.watcher.Add(path)
			}
			return nil
		}
		return filepath.SkipDir
	}

	if base := filepath.Base(path); strings.HasPrefix(base, ".") {
		return nil
	}
	if ext := filepath.Ext(path); ext == ".meta" {
		return nil
	}

	f, err2 := os.Open(path)
	if err2 != nil {
		return err2
	}
	defer f.Close()

	var pol Policy
	if pol, err2 = NewPolicyFromStore(p.language, path, f); err2 != nil {
		return err2
	}

	_, err2 = p.Create(pol)
	return err2
}
