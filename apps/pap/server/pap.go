package server

import (
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

func (s *Services) newPAP() (*pap.PAP, error) {
	opts := []pap.Option{pap.WithLanguage(s.l.Language())}

	db, err := postgresql.New(s.ctx, s.cfg.DatabaseURL, time.Minute*5, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres backend: %w", err)
	}

	opts = append(opts, pap.WithPgPool(db))
	opts = append(opts, pap.WithMigration(s.cfg.DatabaseURL))

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
