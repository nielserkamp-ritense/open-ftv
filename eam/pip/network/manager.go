package network

import (
	"context"
	"log/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/warc"
)

// Manager represents the interface to manage external sources.
type Manager interface{}

// SourceRecorder is invoked for every attribute/entity that an external pull writes.
//
// It lets the PIP register a "Logged" source reference (span_id + WARC file + version)
// for the updated value, without the network package depending on the pip package.
type SourceRecorder func(kind, key, traceID, spanID, warcFile, version string)

// ManagerParams defines the parameters to instantiate a new external sources manager.
type ManagerParams struct {
	Ctx           context.Context          // optional context.
	Path          string                   // required path to config file.
	Logger        *slog.Logger             // required log sink.
	NewAttributes models.AttributesBuilder // required for managing entities and/or relations.
	Attributes    models.AttributeSet      // required for managing attributes.
	Entities      models.EntitySet         // required for managing entities.
	Relations     models.RelationSet       // required for managing relations.
	WARC          *warc.Writer             // optional WARC log for request/response pairs.
	Recorder      SourceRecorder           // optional source-reference recorder.
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
		warc:          params.WARC,
		recorder:      params.Recorder,
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
	attributes    models.AttributeSet
	entities      models.EntitySet
	relations     models.RelationSet
	warc          *warc.Writer
	recorder      SourceRecorder
}
