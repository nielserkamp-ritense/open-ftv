package pap

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// LoadFiles loads all policies from the local file store.
func (p *PAP) LoadFiles() {
	if p.policyStore == "" {
		return
	}

	p.clearWatcher()

	if err := filepath.WalkDir(p.policyStore, p.loadPolicy); err != nil {
		p.logger.Error("pap: error loading policies", "policyStore", p.policyStore, "err", err)
	}
}

func (p *PAP) loadPolicy(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if d.IsDir() {
		if p.recurse || path == p.policyStore {
			if p.policyWatcher != nil {
				_ = p.policyWatcher.Add(path)
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

	var pol *models.Policy
	if pol, err2 = models.NewPolicyFromStore(p.language, path, f); err2 != nil {
		return err2
	}

	_, err2 = p.Create(pol, loadUser)
	return err2
}

const loadUser = "*LOADER*"
