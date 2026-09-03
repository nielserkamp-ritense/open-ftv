package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
)

// NewPAP instantiates a vanilla PAP
func (s *Services) newPAP() (*pap.PAP, error) {
	// Always make the seed file store known so LoadFiles can populate an empty
	// (or freshly migrated) backend from the bundled cedar policy files.
	// NOTE: no file store on this PAP — the embedded PDP's controller calls LoadFiles(),
	// and a postgres-backed PAP must NOT load filename-keyed cedar files (their ids are not
	// UUIDs). The cedar policies are seeded into postgres with UUID ids by seedAuthzPolicies.
	opts := []pap.Option{
		pap.WithLanguage(s.l.Language()),
	}

	opts = append(opts, pap.WithPgPool(s.db))
	opts = append(opts, pap.WithMigration("*embed*", s.cfg.DatabaseURL, true, 0))

	_ = s.cfg.FixTags()

	p, err := pap.New(s.ctx, s.logger, opts...)
	if err != nil {
		return nil, err
	}

	err = p.Migrate()
	if err != nil {
		return nil, fmt.Errorf("pap: migration failed: %w", err)
	}

	return p, nil
}
