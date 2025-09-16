package fiber

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
)

// AuthZENVersion is the full semantic API version for the AuthZEN endpoints.
const AuthZENVersion = "1.0.0"

// AuthZENAuthorizer represents the interface for handling AuthZEN authorization requests.
type AuthZENAuthorizer struct {
	hasEvaluations    bool
	hasSearchSubject  bool
	hasSearchAction   bool
	hasSearchResource bool
	prefix            string
	logger            *slog.Logger
	authLogger        authlog.Logger
	controller        pdp.Controller
}

// NewAuthHandlerZEN instantiates a new AuthZEN authorization handler.
func NewAuthHandlerZEN(logger *slog.Logger, authLogger authlog.Logger, controller pdp.Controller, prefix string) *AuthZENAuthorizer {
	return &AuthZENAuthorizer{
		logger:            logger,
		authLogger:        authLogger,
		controller:        controller,
		prefix:            strings.TrimRight(prefix, "/"),
		hasEvaluations:    false,
		hasSearchSubject:  false,
		hasSearchAction:   false,
		hasSearchResource: false,
	}
}

// WithEvaluations can be used to turn on support for the AuthZEN Evaluations API.
func (h *AuthZENAuthorizer) WithEvaluations() *AuthZENAuthorizer {
	h.hasEvaluations = true
	return h
}

// WithSearchSubject can be used to turn on support for the AuthZEN Subject Search API.
func (h *AuthZENAuthorizer) WithSearchSubject() *AuthZENAuthorizer {
	h.hasSearchAction = true
	return h
}

// WithSearchAction can be used to turn on support for the AuthZEN Action Search API.
func (h *AuthZENAuthorizer) WithSearchAction() *AuthZENAuthorizer {
	h.hasSearchAction = true
	return h
}

// WithSearchResource can be used to turn on support for the AuthZEN Resource Search API.
func (h *AuthZENAuthorizer) WithSearchResource() *AuthZENAuthorizer {
	h.hasSearchResource = true
	return h
}

// Evaluation implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) Evaluation(fc *fiber.Ctx) error {
	prepareResponseHeaders(fc)

	p := initAuthProcess(fc, h.logger, h.authLogger, h.controller)

	if p.logger.Enabled(nil, slog.LevelInfo) {
		defer p.log()
	}
	if p.authLogger != nil {
		defer p.authLog()
	}

	req := p.verifyRequestAuthZEN()
	if p.err != nil {
		p.logger.Error("AuthZEN authorization handler failed", "request", req, "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}

	p.createRequestAuthZEN(req)
	p.logger.Debug("AuthZEN authorization request", "request", p.parc)

	return p.authorizeRequestAuthZEN()
}

// Evaluations implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) Evaluations(fc *fiber.Ctx) error {
	prepareResponseHeaders(fc)

	if !h.hasEvaluations {
		return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
	}

	p := initAuthProcess(fc, h.logger, h.authLogger, h.controller)

	if p.logger.Enabled(nil, slog.LevelInfo) {
		defer p.log()
	}
	if p.authLogger != nil {
		defer p.authLog()
	}

	batch := p.verifyBatchAuthZEN()
	if p.err != nil {
		p.logger.Error("AuthZEN batch authorization handler failed", "request", batch, "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}

	p.createBatchAuthZEN(batch)
	p.logger.Debug("AuthZEN batch authorization request", "request", p.parc)

	return p.authorizeBatchAuthZEN()
}

// SearchSubject implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) SearchSubject(fc *fiber.Ctx) error {
	prepareResponseHeaders(fc)

	if !h.hasSearchSubject {
		return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
	}

	// TODO: implement

	return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
}

// SearchAction implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) SearchAction(fc *fiber.Ctx) error {
	prepareResponseHeaders(fc)

	if !h.hasSearchAction {
		return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
	}

	// TODO: implement

	return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
}

// SearchResource implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) SearchResource(fc *fiber.Ctx) error {
	prepareResponseHeaders(fc)

	if !h.hasSearchResource {
		return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
	}

	// TODO: implement

	return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
}

// Metadata implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) Metadata(fc *fiber.Ctx) error {
	prepareResponseHeaders(fc)

	prefix := fixPrefix(fc, h.prefix)

	out := oas.MetadataResponse{
		PolicyDecisionPoint:      fixDomain(prefix),
		AccessEvaluationEndpoint: prefix + PathEvaluation,
	}

	if h.hasEvaluations {
		out.AccessEvaluationsEndpoint = prefix + PathEvaluations
	}
	if h.hasSearchSubject {
		out.SearchSubjectEndpoint = prefix + PathSearchSubject
	}
	if h.hasSearchAction {
		out.SearchSubjectEndpoint = prefix + PathSearchAction
	}
	if h.hasSearchResource {
		out.SearchSubjectEndpoint = prefix + PathSearchResource
	}

	return fc.JSON(out)
}

func prepareResponseHeaders(fc *fiber.Ctx) {
	fc.Set(HeaderVersion, AuthZENVersion)

	if reqID := fc.Get(HeaderRequestID); reqID != "" {
		fc.Set(HeaderRequestID, reqID)
	}
}

func fixPrefix(fc *fiber.Ctx, prefix string) string {
	if prefix != "" {
		return prefix
	}

	uri := fc.Request().URI()

	path := string(uri.Path())
	switch {
	case strings.HasSuffix(path, PathMetadata):
		path = path[:len(path)-len(PathMetadata)]
	case strings.HasPrefix(path, PathWellKnown), path == "":
		path = PathAuthZEN + PathV1
	}

	return fmt.Sprintf("%s://%s%s", string(uri.Scheme()), string(uri.Host()), path)
}

func fixDomain(prefix string) string {
	uri, _ := url.Parse(prefix)
	return fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)
}
