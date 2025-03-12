package server

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pap/persistence"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
)

func (s *service) newPAP() (pap.PAP, error) {
	opts := make([]pap.Option, 0)
	opts = append(opts, pap.WithLanguage(s.l.Language()))

	if s.cfg.PersistType != "" {
		store, err := persistence.New(s.ctx, s.cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if store != nil {
			opts = append(opts, pap.WithPersistence(store, s.cfg.PersistBase))
		}
	}

	return pap.New(s.ctx, s.logger, opts...), nil
}
