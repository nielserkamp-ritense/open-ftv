package server

import (
	"path/filepath"

	"github.com/gofiber/fiber/v2"

	odrlexport "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/export"
	odrlhandler "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/handler"
	odrlimport "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/importer"
)

// initODRL wires the ODRL-AP-NL import/export endpoints onto the running PDP,
// against the SAME in-process PAP the active engine subscribes to:
//
//	GET  /v1/odrl/export[/{language}/{id}]  export the PAP as ODRL-AP-NL
//	POST /v1/odrl/import                     import an ODRL document (or {"url": ...})
//	POST /v1/odrl/events                     import via CloudEvents 1.0 push
//
// Because the importer writes to the controller's own PAP, a runtime import
// fires the PAP EventSink that the engine registered (see
// odrl-geo/controller.go Handle / opa-embedded/event.go), so the engine reloads
// the imported policy live. This closes finding A2, where the PDP exposed no
// import route and a runtime import never reached the running engine.
//
// NOTE: this synchronises a SINGLE PDP process only. Propagating imports to
// OTHER PDP replicas requires a shared, watch-capable persistence store (e.g.
// etcd) so that a write in one process raises PAP events in the others; that
// shared persist + watch is deliberately out of scope for this fix.
func (s *service) initODRL(group fiber.Router) {
	if s.auth == nil || s.auth.Controller() == nil {
		return
	}
	ap := s.auth.Controller().PAP()
	if ap == nil {
		s.logger.Warn("odrl import/export not registered: controller has no PAP")
		return
	}

	annPath := s.cfg.Annotations
	if annPath == "" && s.cfg.PAP.Store != "" {
		annPath = filepath.Join(s.cfg.PAP.Store, odrlexport.AnnotationsFile)
	}

	ann, err := odrlexport.LoadAnnotations(annPath)
	if err != nil {
		s.logger.Error("failed to load ODRL export annotations", "path", annPath, "error", err)
		ann = nil
	}

	exporter := odrlexport.New(ap, ann, s.cfg.BaseURL)
	imp := odrlimport.New(ap, s.logger, nil)

	h := odrlhandler.New(s.logger, ap, exporter, imp)
	h.Register(group)
	s.logger.Info("odrl import/export endpoints registered on pdp (in-process PAP sync)")
}
