package server

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"

	odrlevents "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/events"
	odrlexport "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/export"
	odrlhandler "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/handler"
	odrlimport "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/importer"
)

// initODRL wires the ODRL-AP-NL im-/export endpoints onto the PAP:
// GET /v1/odrl/export[/{language}/{id}], POST /v1/odrl/import and
// POST /v1/odrl/events. When subscribers and/or an export directory are
// configured, an event emitter is registered on the PAP as well.
func (s *service) initODRL(group fiber.Router) {
	cfg := s.cfg.Config // the embedded odrl.Config.

	annPath := cfg.Annotations
	if annPath == "" && s.cfg.PAP.Store != "" {
		annPath = filepath.Join(s.cfg.PAP.Store, odrlexport.AnnotationsFile)
	}

	ann, err := odrlexport.LoadAnnotations(annPath)
	if err != nil {
		s.logger.Error("failed to load ODRL export annotations", "path", annPath, "error", err)
		ann = nil
	}

	exporter := odrlexport.New(s.pap, ann, cfg.BaseURL)
	imp := odrlimport.New(s.pap, s.logger, nil)

	h := odrlhandler.New(s.logger, s.pap, exporter, imp)
	h.Register(group)

	if subscribers := cfg.SubscriberList(); len(subscribers) > 0 || cfg.ExportDir != "" {
		emitter := odrlevents.New(s.ctx, s.logger, s.pap, exporter, subscribers, cfg.ExportDir, cfg.Source)
		s.pap.AddEventSink(emitter)
		s.logger.Info("odrl event emitter registered", "subscribers", len(subscribers), "exportDir", cfg.ExportDir)
	}
}
