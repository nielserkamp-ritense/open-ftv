package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/persistence"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
)

func (s *service) newPIP() (pip.PIP, error) {
	opts := make([]pip.Option, 0)

	if s.cfg.PersistType != "" {
		store, err := persistence.New(s.ctx, s.cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if store != nil {
			opts = append(opts, pip.WithPersistence(store, s.cfg.PersistBase))
		}
	}

	return pip.New(s.ctx, s.logger, opts...), nil
}
