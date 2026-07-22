package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// NewPAP instantiates a vanilla PAP with an optional persistent storage backend.
func (s *Services) newPAP() (*pap.PAP, error) {
	// Always make the seed file store known so LoadFiles can populate an empty
	// (or freshly migrated) backend from the bundled cedar policy files.
	// NOTE: no file store on this PAP — the embedded PDP's controller calls LoadFiles(),
	// and a postgres-backed PAP must NOT load filename-keyed cedar files (their ids are not
	// UUIDs). The cedar policies are seeded into postgres with UUID ids by seedAuthzPolicies.
	opts := []pap.Option{
		pap.WithLanguage(s.l.Language()),
	}

	switch {
	case s.db != nil:
		opts = append(opts, pap.WithPgPool(s.db))
	case s.store != nil:
		base := fmt.Sprintf("%s%s%s", convert.ForceSuffix(s.cfg.Base, pap.PathSeparator), "policies", pap.PathSeparator)
		opts = append(opts, pap.WithKeyValueDB(s.store, base))
	}

	if s.cfg.Source != "" {
		switch {
		case s.db != nil:
			opts = append(opts, pap.WithMigration(s.cfg.Source, s.cfg.Persist.PgURL, s.cfg.Auto, s.cfg.Steps))
		}
	}

	_ = s.cfg.FixTags()

	p := pap.New(s.ctx, s.logger, opts...)
	if p == nil {
		return nil, fmt.Errorf("failed to create pap")
	}

	return p, nil
}
