package fiber

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// AuthZENVersion is the full semantic API version for the AuthZEN endpoints.
const AuthZENVersion = "1.4.0"

// AuthZENAuthorizer represents the interface for handling AuthZEN authorization requests.
type AuthZENAuthorizer struct {
	hasEvaluations    bool
	hasSearchSubject  bool
	hasSearchAction   bool
	hasSearchResource bool
	prefix            string
	logger            *slog.Logger
	adl               *adl.ADL
	controller        pdp.Controller
}

// NewAuthHandlerZEN instantiates a new AuthZEN authorization handler.
func NewAuthHandlerZEN(logger *slog.Logger, decisionLog *adl.ADL, controller pdp.Controller, prefix string) *AuthZENAuthorizer {
	return &AuthZENAuthorizer{
		logger:     logger,
		adl:        decisionLog,
		controller: controller,
		prefix:     strings.TrimRight(prefix, "/"),
	}
}

// WithEvaluations can be used to turn on support for the AuthZEN Evaluations API.
func (h *AuthZENAuthorizer) WithEvaluations() *AuthZENAuthorizer {
	h.hasEvaluations = true
	return h
}

// WithSearchSubject can be used to turn on support for the AuthZEN Subject Search API.
func (h *AuthZENAuthorizer) WithSearchSubject() *AuthZENAuthorizer {
	h.hasSearchSubject = true
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

// Metadata implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) Metadata(fc *fiber.Ctx) error {
	processHeaders(fc)

	prefix := fixPrefix(fc, h.prefix)
	domain := fixDomain(prefix)

	out := oas.MetadataResponse{
		PolicyDecisionPoint:      domain,
		AccessEvaluationEndpoint: prefix + PathEvaluation,
	}

	if h.hasEvaluations {
		out.AccessEvaluationsEndpoint = prefix + PathEvaluations
	}
	if h.hasSearchSubject {
		out.SearchSubjectEndpoint = prefix + PathSearchSubject
	}
	if h.hasSearchAction {
		out.SearchActionEndpoint = prefix + PathSearchAction
	}
	if h.hasSearchResource {
		out.SearchResourceEndpoint = prefix + PathSearchResource
	}

	return fc.JSON(out)
}

func processHeaders(fc *fiber.Ctx) {
	processTraceParent(fc)
	processFSCTransactionID(fc)
	prepareResponseHeaders(fc)
}

// processFSCTransactionID stashes the Fsc-Transaction-Id header (adl.fsc.transaction_id, §3.3.7.6) on context.
func processFSCTransactionID(fc *fiber.Ctx) {
	txnID := strings.TrimSpace(fc.Get(models.HeaderFSCTransactionID))
	if txnID == "" {
		return
	}

	fc.SetUserContext(context.WithValue(fc.UserContext(), models.AttrFSCTransactionID, txnID))
}

// processTraceParent applies the header only; no fallback here, so applyTraceFallback can still prefer
// a body-embedded traceparent (Logius ADL §4.1.1.1 Example 10) once the request body is parsed.
func processTraceParent(fc *fiber.Ctx) {
	incoming := strings.TrimSpace(fc.Get(models.HeaderTraceParent))
	if incoming == "" {
		return
	}

	if tc, ok := models.ParseTraceParent(incoming); ok {
		applyDecisionTrace(fc, tc.TraceID, tc.SpanID, models.NewSpanID())
	}
}

func applyDecisionTrace(fc *fiber.Ctx, traceID, parentSpanID, spanID string) {
	ctx := fc.UserContext()
	ctx = context.WithValue(ctx, models.AttrTraceID, traceID)
	ctx = context.WithValue(ctx, models.AttrSpanID, spanID)

	if parentSpanID != "" {
		ctx = context.WithValue(ctx, models.AttrParentSpanID, parentSpanID)
	}

	fc.SetUserContext(ctx)
}

func applyTraceParent(fc *fiber.Ctx, traceParent string) {
	tc, ok := models.ParseTraceParent(traceParent)
	if !ok {
		return
	}

	applyDecisionTrace(fc, tc.TraceID, tc.SpanID, models.NewSpanID())
}

// applyTraceFallback guarantees every request ends up with a trace/span pair by the time ADL logging happens.
// Logius ADL sets the precedence: the HTTP traceparent header (already applied by processTraceParent) wins.
// When that's absent, a traceparent embedded in the AuthZEN request body's context object is used instead
// (Logius ADL §4.1.1.1 Example 10).
// If neither is present, a fresh trace/span pair is generated so the ADL record still gets one.
func applyTraceFallback(fc *fiber.Ctx, attrs *models.AttributeSet) {
	if convert.AnyToString(fc.UserContext().Value(models.AttrTraceID)) != "" {
		return
	}

	if attrs != nil {
		if tp := convert.AnyToString(attrs.GetAttributeValue(models.AttrTraceParent)); tp != "" {
			applyTraceParent(fc, models.ResolveTraceParent(tp))
			return
		}
	}

	_, tc := models.NewTraceParent()
	applyDecisionTrace(fc, tc.TraceID, "", tc.SpanID)
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

// reasonContext builds the AuthZEN response's context object: "id" set from the decision outcome,
// and the human-readable reason in reason_user, falling back to a fixed message when the
// controller didn't supply one.
func reasonContext(resp *models.Response) oas.ReasonObject {
	id, msg := "ok", resp.Message

	if !resp.Allowed {
		id = "not-authorized"

		if msg == "" {
			msg = "not authorized"
		}
	} else if msg == "" {
		msg = "ok"
	}

	return oas.ReasonObject{Id: id, ReasonUser: oas.ReasonField{"en": msg}}
}
