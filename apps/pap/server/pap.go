package server

import (
	"fmt"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
)

func (s *service) newPAP() (pap2.PAP, error) {
	opts := make([]pap2.Option, 0)
	opts = append(opts, pap2.WithLanguage(s.l.Language()))

	if s.cfg.Persist.Type != "" {
		store, err := s.cfg.Persist.NewStore(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if store != nil {
			opts = append(opts, pap2.WithPersistence(store, s.cfg.Persist.Base))
		}
	}

	return pap2.New(s.ctx, s.logger, opts...), nil
}
