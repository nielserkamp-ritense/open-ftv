package server

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// seedNamespace makes the seeded policy ids deterministic (stable across restarts).
var seedNamespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte("openftv-mgmt-authz-policies"))

// seedAuthzPolicies seeds the bundled cedar authorization policies (admin/author/auditor/…)
// into the PAP store with deterministic UUID ids, so the manager's embedded PDP enforces
// them AND they are visible/editable in the UI (the postgres policy table requires UUID
// ids, so the filename-keyed seed files cannot be loaded directly). It only seeds when the
// store is empty; once seeded, the store (postgres) is the source of truth and UI edits win.
func (s *Services) seedAuthzPolicies() {
	if s.pap == nil {
		return
	}

	list, err := s.pap.List("")
	if err != nil {
		s.logger.Error("seed: failed to list policies", "err", err)
		return
	}
	if len(list) > 0 {
		s.logger.Info("seed: policy store already populated; leaving it as the source of truth", "count", len(list))
		return
	}

	dir := s.cfg.PAP.Store
	entries, err := os.ReadDir(dir)
	if err != nil {
		s.logger.Error("seed: cannot read policy seed dir", "dir", dir, "err", err)
		return
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cedar") {
			continue
		}

		content, rerr := os.ReadFile(filepath.Join(dir, e.Name()))
		if rerr != nil {
			s.logger.Error("seed: cannot read policy file", "file", e.Name(), "err", rerr)
			continue
		}

		id := uuid.NewSHA1(seedNamespace, []byte(e.Name())).String()
		pol, perr := models.NewPolicyFromData(id, "cedar", "", "", bytes.NewReader(content))
		if perr != nil {
			s.logger.Error("seed: cannot build policy", "file", e.Name(), "err", perr)
			continue
		}

		// The policy store has a UNIQUE(language, title) index (policy_ix1); without a
		// distinct title every seeded cedar policy collides and only the first is stored,
		// leaving the role policies missing. Use the file's base name as a stable title.
		pol = pol.WithTitle(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))

		if _, cerr := s.pap.Create(pol, "*SEED*"); cerr != nil {
			s.logger.Error("seed: cannot create policy", "file", e.Name(), "err", cerr)
			continue
		}
		s.logger.Info("seed: authorization policy seeded into store", "file", e.Name(), "id", id)
	}
}
