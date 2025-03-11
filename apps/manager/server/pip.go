package server

import (
	"context"
	"fmt"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/manager/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/manager/persistence"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// NewPIP instantiates a vanilla PIP with an optional persistent storage backend.
func NewPIP(ctx context.Context, cfg *config.Config, logger *slog.Logger, l models.Language) (pip.PIP, error) {
	opts := []pip.Option{pip.WithFileStore(cfg.PipStore, cfg.PipStoreRecurse), pip.WithPullConfigs(cfg.PipPullConfigs)}
	if l == models.CEDAR {
		opts = append(opts, pip.WithFactories(cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)))
	}

	if cfg.PersistType != "" {
		s, err := persistence.New(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create persistence store: %w", err)
		}
		if s != nil {
			base := convert.ForceSuffix(cfg.PersistBase, "/") + "data/"
			opts = append(opts, pip.WithPersistence(s, base))
		}
	}

	return pip.New(ctx, logger, opts...), nil
}
