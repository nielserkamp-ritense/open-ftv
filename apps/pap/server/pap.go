package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
)

func (s *service) newPAP() (pap.PAP, error) {
	opts := make([]pap.Option, 0)
	opts = append(opts, pap.WithLanguage(s.l.Language()))

	if s.cfg.Persist.Type != "" {
		store, err := s.cfg.Persist.NewStore(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if store != nil {
			opts = append(opts, pap.WithPersistence(store, s.cfg.Persist.Base))
		}
	}

	return pap.New(s.ctx, s.logger, opts...), nil
}
