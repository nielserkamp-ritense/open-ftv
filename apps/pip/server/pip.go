package server

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pip/persistence"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
)

// NewPIP instantiates a vanilla PIP with an optional persistent storage backend.
func NewPIP(ctx context.Context, cfg *config.Config, logger *slog.Logger) (pip.PIP, error) {
	opts := make([]pip.Option, 0)

	if cfg.PersistType != "" {
		s, err := persistence.New(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if s != nil {
			opts = append(opts, pip.WithPersistence(s, cfg.PersistBase))
		}
	}

	return pip.New(ctx, logger, opts...), nil
}
