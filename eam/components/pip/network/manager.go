package network

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Manager represents the interface to manage external sources.
type Manager interface{}

// ManagerParams defines the parameters to instantiate a new external sources manager.
type ManagerParams struct {
	Ctx           context.Context          // optional context.
	Path          string                   // required path to config file.
	Logger        *slog.Logger             // required log sink.
	NewAttributes models.AttributesBuilder // required for managing entities and/or relations.
	Attributes    models.AttributeSet      // required for managing attributes.
	Entities      models.EntitySet         // required for managing entities.
	Relations     models.RelationSet       // required for managing relations.
}

// NewManager instantiates a new external sources manager.
func NewManager(params ManagerParams) (Manager, error) {
	cfg, err := LoadConfig(params.Path)
	if err != nil {
		params.Logger.Error("failed to load manager configuration", "path", params.Path, "error", err)
		return nil, err
	}

	m := &manager{
		cfg:           cfg,
		logger:        params.Logger,
		newAttributes: params.NewAttributes,
		attributes:    params.Attributes,
		entities:      params.Entities,
		relations:     params.Relations,
	}

	if params.Ctx == nil {
		params.Ctx = context.Background()
	}

	m.ctx, m.cancel = context.WithCancel(params.Ctx)

	go m.schedule()
	return m, nil
}

type manager struct {
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *slog.Logger
	cfg           *Config
	newAttributes models.AttributesBuilder
	attributes    models.AttributeSet
	entities      models.EntitySet
	relations     models.RelationSet
}
