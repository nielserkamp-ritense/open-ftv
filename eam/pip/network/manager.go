package network

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Manager represents the interface to manage external sources.
type Manager interface{}

// ManagerParams defines the parameters to instantiate a new external sources manager.
type ManagerParams struct {
	Ctx           context.Context          // optional context.
	Path          string                   // required path to config-file.
	Logger        *slog.Logger             // required log sink.
	NewAttributes models.AttributesBuilder // required for managing entities and/or relations.
	AddAttribute  models.AddAttribute      // required for adding attributes.
	GetAttribute  models.GetAttribute      // required for retrieving attribute values.
	AddEntity     models.AddEntity         // required for adding entities.
	AddRelation   models.AddRelation       // required for adding relations.
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
		addAttribute:  params.AddAttribute,
		getAttribute:  params.GetAttribute,
		addEntity:     params.AddEntity,
		addRelation:   params.AddRelation,
	}

	ctx := params.Ctx
	if params.Ctx == nil {
		ctx = context.Background()
	}

	m.ctx, m.cancel = context.WithCancel(ctx)

	go m.schedule()
	return m, nil
}

type manager struct {
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *slog.Logger
	cfg           *Config
	newAttributes models.AttributesBuilder
	addAttribute  models.AddAttribute
	getAttribute  models.GetAttribute
	addEntity     models.AddEntity
	addRelation   models.AddRelation
}
