package server

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// seedNamespace makes the seeded policy ids deterministic (stable across restarts).
var seedNamespace = uuid.NewSHA1(uuid.NameSpaceURL, []byte("openftv-mgmt-authz-policies"))

var seedUser = identity.NewSeedPrincipal()

// seedLanguageByExt maps a policy file's extension to the language string
// expected by models.NewPolicyFromData. Only languages actually authored as
// seed files in this repo need an entry here.
var seedLanguageByExt = map[string]string{
	".cedar": "cedar",
	".rego":  "rego",
}

// seedAuthzPolicies seeds the bundled authorization policies (admin/author/auditor/…,
// one file per seedLanguageByExt entry) into the PAP store with deterministic UUID ids,
// so the manager's embedded PDP enforces them AND they are visible/editable in the UI
// (the postgres policy table requires UUID ids, so the filename-keyed seed files cannot
// be loaded directly). It only seeds when the store is empty; once seeded, the store
// (postgres) is the source of truth and UI edits win.
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
		if e.IsDir() {
			continue
		}

		language, ok := seedLanguageByExt[filepath.Ext(e.Name())]
		if !ok {
			continue
		}

		policyPath := filepath.Join(dir, e.Name())

		content, rerr := os.ReadFile(policyPath)
		if rerr != nil {
			s.logger.Error("seed: cannot read policy file", "file", e.Name(), "err", rerr)
			continue
		}

		id := uuid.NewSHA1(seedNamespace, []byte(e.Name())).String()

		pol, perr := models.NewPolicyFromData(id, language, "", "", bytes.NewReader(content))
		if perr != nil {
			s.logger.Error("seed: cannot build policy", "file", e.Name(), "err", perr)
			continue
		}

		// The policy store has a UNIQUE(language, title) index (policy_ix1); without a
		// distinct title every seeded policy collides with same-language siblings and only
		// the first is stored. Use the file's base name as a stable title.
		pol = pol.WithTitle(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))

		if tags := readSeedMetaTags(policyPath); len(tags) > 0 {
			pol = pol.WithTags(tags...)
		}

		if _, cerr := s.pap.Create(pol, seedUser); cerr != nil {
			s.logger.Error("seed: cannot create policy", "file", e.Name(), "err", cerr)
			continue
		}

		s.logger.Info("seed: authorization policy seeded into store", "file", e.Name(), "id", id)
	}
}

func readSeedMetaTags(policyPath string) []string {
	for _, metaPath := range []string{
		strings.TrimSuffix(policyPath, filepath.Ext(policyPath)) + ".meta",
		policyPath + ".meta",
	} {
		data, err := os.ReadFile(metaPath)
		if err != nil {
			continue
		}

		var sidecar oas.Policy
		if err = yaml.Unmarshal(data, &sidecar); err != nil {
			if err = json.Unmarshal(data, &sidecar); err != nil {
				continue
			}
		}

		if len(sidecar.Metadata.Tags) > 0 {
			return sidecar.Metadata.Tags
		}
	}

	return nil
}
