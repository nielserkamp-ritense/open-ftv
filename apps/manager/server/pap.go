package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
)

// newPAP instantiates a postgres-backed PAP and migrates its database.
func (s *Services) newPAP() (*pap.PAP, error) {
	// NOTE: no file store on this PAP — the embedded PDP's controller calls LoadFiles(),
	// and a postgres-backed PAP must NOT load filename-keyed cedar files (their ids are not
	// UUIDs). The cedar policies are seeded into postgres with UUID ids by seedAuthzPolicies.
	opts := []pap.Option{
		pap.WithLanguage(s.l.Language()),
		pap.WithPgPool(s.db),
		pap.WithMigration(s.cfg.DatabaseURL),
	}

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
