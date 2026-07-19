package pap

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
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

// LoadStore replays a PolicyAdded event for every policy currently in the backing store,
// so subscribers (e.g. an embedded PDP) rebuild their policy set from the persisted store.
// Used when policies already live in the store (e.g. a postgres-backed store at startup),
// where LoadFiles is a no-op because there is no local file store.
func (p *PAP) LoadStore() {
	list, err := p.List("")
	if err != nil {
		p.logger.Error("pap: error loading policies from store", "err", err)
		return
	}
	for _, pol := range list {
		if p.eventSinks != nil {
			p.sendEvent(models.PolicyAdded, pol.Key())
		}
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

	// The backing store is the source of truth. If this policy already exists
	// (e.g. a postgres-backed store on restart, possibly carrying UI edits), do
	// NOT re-create it from the seed file — that would version-bump/duplicate it
	// and clobber UI changes. Instead re-emit an event for the STORED policy so
	// subscribers (the embedded PDP controller) rebuild their policy set from the
	// persisted content. On a fresh/empty store the read misses and we seed via
	// Create, preserving the original file-load behavior.
	if existing, _, rerr := p.Read(pol.ID()); rerr == nil && existing != nil {
		if p.eventSinks != nil {
			p.sendEvent(models.PolicyReplaced, existing.Key())
		}
		return nil
	}

	_, err2 = p.Create(pol, loadUser)
	return err2
}

// LoadTags loads a list of standard tags into the database.
func (p *PAP) LoadTags(tags []*policies.Tag) error {
	if p.tagDB == nil {
		return errors.New("pap: tagDB not initialized")
	}
	return p.tagDB.EnsureTags(tags, loadUser)
}

var loadUser = identity.NewLoaderPrincipal()
