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
// be loaded directly).
//
// Seeding is decided per file, not per store: a file whose id is absent is inserted, a file
// whose id is present is left alone, and a file whose id was present once and has since been
// deleted is not brought back. That way a release that adds a seed file reaches a store that
// was seeded by an earlier release, while the store stays the source of truth for every row
// the operator has touched. See docs/adr/0006-management-plane-authorization-model.md.
//
// Telling "never seeded" from "seeded and then deleted" needs the audit trail, which only
// the postgres store keeps. Any other store (etcd, consul) is seeded only while it is
// empty, as before: a file a later release adds does not reach such a store, but a file
// the operator deleted never comes back either.
func (s *Services) seedAuthzPolicies() {
	if s.pap == nil {
		return
	}

	if s.db == nil {
		list, err := s.pap.List("")
		if err != nil {
			s.logger.Error("seed: failed to list policies", "err", err)
			return
		}

		if len(list) > 0 {
			s.logger.Info("seed: policy store already populated and keeps no audit trail; leaving it as the source of truth", "count", len(list))
			return
		}
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

		s.seedPolicyFile(filepath.Join(dir, e.Name()), language)
	}
}

// seedPolicyFile inserts one seed file into the store, unless the store has seen it before.
func (s *Services) seedPolicyFile(policyPath, language string) {
	name := filepath.Base(policyPath)
	id := uuid.NewSHA1(seedNamespace, []byte(name)).String()

	if seeded, err := s.alreadySeeded(id); err != nil {
		s.logger.Error("seed: cannot inspect policy store", "file", name, "id", id, "err", err)
		return
	} else if seeded {
		s.logger.Debug("seed: policy present or removed by the operator; leaving the store as the source of truth", "file", name, "id", id)
		return
	}

	content, err := os.ReadFile(policyPath)
	if err != nil {
		s.logger.Error("seed: cannot read policy file", "file", name, "err", err)
		return
	}

	pol, err := models.NewPolicyFromData(id, language, "", "", bytes.NewReader(content))
	if err != nil {
		s.logger.Error("seed: cannot build policy", "file", name, "err", err)
		return
	}

	// The policy store has a UNIQUE(language, title) index (policy_ix1); without a
	// distinct title every seeded policy collides with same-language siblings and only
	// the first is stored. Use the file's base name as a stable title.
	pol = pol.WithTitle(strings.TrimSuffix(name, filepath.Ext(name)))

	if tags := readSeedMetaTags(policyPath); len(tags) > 0 {
		pol = pol.WithTags(tags...)
	}

	if _, err = s.pap.Create(pol, seedUser); err != nil {
		s.logger.Error("seed: cannot create policy", "file", name, "err", err)
		return
	}

	s.logger.Info("seed: authorization policy seeded into store", "file", name, "id", id)
}

// alreadySeeded reports whether the seed file with the given id has ever reached the store.
//
// A present row is the obvious case. A row that is absent but has an audit trail was seeded
// earlier and then deleted by the operator, which ADR 0001 accepts as their call; the audit
// table keeps that trail because it carries no foreign key to the policy (ADR 0005).
func (s *Services) alreadySeeded(id string) (bool, error) {
	if pol, _, err := s.pap.Read(id); err != nil {
		return false, err
	} else if pol != nil {
		return true, nil
	}

	audit, err := s.pap.ReadAudit(id)
	if err != nil {
		return false, err
	}

	return len(audit) > 0, nil
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
