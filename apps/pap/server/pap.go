package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
)

func (s *Services) newPAP() (*pap.PAP, error) {
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
