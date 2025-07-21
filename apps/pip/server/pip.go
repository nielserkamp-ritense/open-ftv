package server

import (
	"fmt"

	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

func (s *service) newPIP() (*pip2.PIP, error) {
	opts := make([]pip2.Option, 0)

	if s.cfg.Persist.Type != "" {
		store, err := s.cfg.Persist.NewStore(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if store != nil {
			opts = append(opts, pip2.WithPersistence(store, s.cfg.Persist.Base))
		}
	}

	return pip2.New(s.ctx, s.logger, opts...), nil
}
