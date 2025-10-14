package server

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

func (s *Services) newPIP() (*pip.PIP, error) {
	opts := make([]pip.Option, 0)
	isPG := strings.EqualFold(s.cfg.Persist.Type, "postgres")

	switch {
	case isPG:
		db, err := pip.NewPostgresDB(s.ctx, s.cfg.Persist.PgURL, s.cfg.Persist.PgMaxLife, s.cfg.PgMaxConn)
		if err != nil {
			return nil, fmt.Errorf("failed to connect postgres backend: %w", err)
		}
		opts = append(opts, pip.WithPostgresDB(db))

	case s.cfg.Persist.Type != "":
		store, err := s.cfg.Persist.NewStore(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if store != nil {
			opts = append(opts, pip.WithKeyValueDB(store, s.cfg.Persist.Base))
		}
	}

	return pip.New(s.ctx, s.logger, opts...), nil
}
