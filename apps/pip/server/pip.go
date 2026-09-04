package server

import (
	"fmt"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
)

func (s *Services) newPIP() (*pip.PIP, error) {
	opts := make([]pip.Option, 0)

	db, err := pip.NewPostgresDB(s.ctx, s.cfg.DatabaseURL, 5*time.Minute, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to connect postgres backend: %w", err)
	}

	opts = append(opts, pip.WithPostgresDB(db))

	return pip.New(s.ctx, s.logger, opts...)
}
