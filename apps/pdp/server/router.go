package server

import (
	"context"

	"github.com/gofiber/fiber/v2"

	handle "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// initRoutes sets up the routing table for HTTP requests.
func (s *service) initRoutes(ctx context.Context, svc *fiber.App) {
	s.ctx = ctx
	s.l = models.LanguageFromString(s.cfg.PAP.Language)

	s.auth = s.newAuth(handle.PathAuthZEN + handle.PathV1)
	if s.auth == nil {
		panic("failed to initialize authorization handler")
	}

	v1 := svc.Group(handle.PathV1)

	s.initHealth(svc)
	s.initAuth(svc)
	s.initBundles(v1)
}

func (s *service) initHealth(svc *fiber.App) {
	// liveness and readiness.
	svc.Get(handle.PathHealthZ, s.chk.HealthZ)
	svc.Get(handle.PathLiveZ, s.chk.LiveZ)
	svc.Get(handle.PathReadyZ, s.chk.ReadyZ)
}

func (s *service) initAuth(svc *fiber.App) {
	// AuthZEN
	authZen := svc.Group(handle.PathAuthZEN)
	authZenV1 := authZen.Group(handle.PathV1)
	authZenV1.Post(handle.PathEvaluation, s.auth.zen.Evaluation)
	authZenV1.Post(handle.PathEvaluations, s.auth.zen.Evaluations)
	// authZenV1.Post(handle.PathSearchSubject, s.auth.zen.SearchSubject)
	// authZenV1.Post(handle.PathSearchAction, s.auth.zen.SearchAction)
	// authZenV1.Post(handle.PathSearchResource, s.auth.zen.SearchResource)
	authZenV1.Get(handle.PathMetadata, s.auth.zen.Metadata)

	// Metadata also available on /.well-known/authzen-configuration.
	wellKnown := svc.Group(handle.PathWellKnown)
	wellKnown.Get(handle.PathAuthZenConfig, s.auth.zen.Metadata)
}

func (s *service) initBundles(v1 fiber.Router) {
	handler := handle.NewBundleReceiverHandler(s.logger, s.auth.controller, s.auth.authorizer)
	v1.Post(handle.PathBundle, handler.PostBundle)

	if url := s.cfg.PDP.BundleManager; url != "" {
		go s.bundleRetriever(url, handler)
	}
}

func (s *service) bundleRetriever(url string, handler *handle.BundleReceiverHandler) {
	s.logger.Info("retrieving latest bundle from manager", "url", url)

	for {
		resp, err := s.getLatestBundle(url)
		if err != nil {
			s.logger.Error("failed to get latest bundle", "url", url, "error", err)
			continue
		}

		ct := resp.Header.Get(fiber.HeaderContentEncoding)
		if _, err = handler.ProcessBundle(ct, resp.Body); err != nil {
			resp.Body.Close()
			s.logger.Error("failed to process latest bundle", "url", url, "error", err)
			continue
		}

		resp.Body.Close()
		break
	}

	s.chk.SetReady(true) // now we can handle requests!
}
