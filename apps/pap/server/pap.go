package server

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

func (s *Services) newPAP() (*pap.PAP, error) {
	opts := []pap.Option{pap.WithLanguage(s.l.Language())}
	isPG := strings.EqualFold(s.cfg.Persist.Type, "postgres")

	switch {
	case isPG:
		db, err := postgresql.New(s.ctx, s.cfg.PgURL, s.cfg.PgMaxLife, s.cfg.PgMaxConn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect postgres backend: %w", err)
		}

		opts = append(opts, pap.WithPgPool(db))

	case s.cfg.Persist.Type != "":
		store, err := s.cfg.NewStore(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}

		if store != nil {
			opts = append(opts, pap.WithKeyValueDB(store, s.cfg.Base))
		}
	}

	if s.cfg.Source != "" {
		switch {
		case isPG:
			opts = append(opts, pap.WithMigration(s.cfg.Source, s.cfg.PgURL, s.cfg.Auto, s.cfg.Steps))
		}
	}

	_ = s.cfg.FixTags()

	p, err := pap.New(s.ctx, s.logger, opts...)
	if err != nil {
		return nil, err
	}

	if err = p.Migrate(); err != nil {
		return nil, fmt.Errorf("pap: migration failed: %w", err)
	}

	return p, nil
}
