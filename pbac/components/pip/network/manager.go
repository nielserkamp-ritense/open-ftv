package network

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

// Manager represents the interface to manage external sources.
type Manager interface {
}

// NewManager instantiates a new external sources manager.
func NewManager(ctx context.Context, path string, logger *slog.Logger, get models.GetAttribute) (Manager, error) {
	cfg, err := LoadConfig(path)
	if err != nil {
		logger.Error("failed to load manager configuration", "path", path, "error", err)
		return nil, err
	}

	m := &manager{cfg: cfg, logger: logger, get: get}
	m.ctx, m.cancel = context.WithCancel(ctx)

	go m.schedule()
	return m, nil
}

type manager struct {
	ctx    context.Context
	cancel context.CancelFunc
	logger *slog.Logger
	cfg    *Config
	get    models.GetAttribute
}
