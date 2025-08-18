package server

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
)

func (s *service) newPAP() (*pap.PAP, error) {
	opts := []pap.Option{pap.WithLanguage(s.l.Language())}
	isPG := strings.EqualFold(s.cfg.Persist.Type, "postgres")

	switch {
	case isPG:
		db, err := pap.NewPostgresDB(s.ctx, s.cfg.Persist.PgURL, s.cfg.Persist.PgMaxLife, s.cfg.PgMaxConn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect postgres backend: %w", err)
		}
		opts = append(opts, pap.WithPostgresDB(db))

	case s.cfg.Persist.Type != "":
		store, err := s.cfg.Persist.NewStore(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if store != nil {
			opts = append(opts, pap.WithKeyValueDB(store, s.cfg.Persist.Base))
		}
	}

	if s.cfg.Migration.Source != "" {
		switch {
		case isPG:
			opts = append(opts, pap.WithMigration(s.cfg.Migration.Source, s.cfg.Persist.PgURL, s.cfg.Migration.Auto, s.cfg.Migration.Steps))
		}
	}

	p := pap.New(s.ctx, s.logger, opts...)
	if p == nil {
		return nil, fmt.Errorf("failed to create pap")
	}
	return p, nil
}
