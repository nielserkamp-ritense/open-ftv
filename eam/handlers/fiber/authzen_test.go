package fiber

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	cedar_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller/adl"
	opa_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	otel "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestAuthHandler_AuthZEN1(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_update","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"ok","reason_user":{"en":"ok"}},"decision":true}`

	t.Run("authzen handler (1)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, ip)

		ap := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, ap)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller, "http://localhost/v1")
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen", buf)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, AuthZENVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 9, h.Count())
	})
}

func TestAuthHandler_AuthZEN2(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"ok","reason_user":{"en":"ok"}},"decision":true}`

	t.Run("authzen handler (2)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, ip)

		ap := pap.New(ctx, logger, pap.WithLanguage("rego"), pap.WithFileStore("../../../testdata/unittest/opa", true))
		require.NotNil(t, ap)

		controller := opa_embedded.NewController(pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller, "http://localhost/v1")
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen", buf)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, AuthZENVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 11, h.Count())
	})
}

func TestAuthHandler_AuthZEN3(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"ok","organisation_id":"12345","reason_user":{"en":"ok"},"scope":["read"]},"decision":true}`

	t.Run("authzen handler, policy context spread across the context object", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, ip)

		ap := pap.New(ctx, logger, pap.WithLanguage("rego"), pap.WithFileStore("../../../testdata/unittest/opa3", true))
		require.NotNil(t, ap)

		controller := opa_embedded.NewController(pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller, "http://localhost/v1")
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen", buf)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, AuthZENVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
	})
}

func TestReasonContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		resp models.Response
		want string
	}{
		{
			name: "no context",
			resp: models.Response{Allowed: true},
			want: `{"id":"ok","reason_user":{"en":"ok"}}`,
		},
		{
			name: "context spread across the context object",
			resp: models.Response{Allowed: true, Context: map[string]any{"organisation_id": "12345", "scope": []any{"read"}}},
			want: `{"id":"ok","organisation_id":"12345","reason_user":{"en":"ok"},"scope":["read"]}`,
		},
		{
			name: "context on a denied decision",
			resp: models.Response{Context: map[string]any{"organisation_id": "12345"}},
			want: `{"id":"not-authorized","organisation_id":"12345","reason_user":{"en":"not authorized"}}`,
		},
		{
			name: "context cannot shadow the AuthZEN properties",
			resp: models.Response{Allowed: true, Context: map[string]any{"id": "spoofed", "reason_user": map[string]any{"en": "spoofed"}, "reason_admin": "spoofed"}},
			want: `{"id":"ok","reason_user":{"en":"ok"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data, err := json.Marshal(reasonContext(&tc.resp))
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(data))
		})
	}
}

func TestAuthHandler_AuthZEN_Fail1(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"GET"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"not-authorized","reason_user":{"en":"not authorized"}},"decision":false}`

	t.Run("authzen handler fail (1)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, ip)

		ap := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, ap)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller, "http://localhost/v1")
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen", buf)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, AuthZENVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 10, h.Count())
	})
}

func TestAuthHandler_AuthZEN_Fail2(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"detail":"invalid subject","status":400,"title":"Bad Request"}`

	t.Run("authzen handler fail (2)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller, "http://localhost/v1")
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen", buf)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, AuthZENVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 7, h.Count())
	})
}

func TestAuthHandler_AuthZEN_Fail3(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"detail":"invalid action","status":400,"title":"Bad Request"}`

	t.Run("authzen handler fail (3)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller, "http://localhost/v1")
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen", buf)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, AuthZENVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 7, h.Count())
	})
}

func TestAuthHandler_AuthZEN_Fail4(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service"}}`
	out := `{"detail":"invalid resource","status":400,"title":"Bad Request"}`

	t.Run("authzen handler fail (4)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, ip)

		ap := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, ap)

		controller := cedar_embedded.NewController(pdp.WithPEP(ep), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller, "http://localhost/v1")
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen", buf)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		assert.Equal(t, AuthZENVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 7, h.Count())
	})
}

func TestProcessFSCTransactionID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		header string
		want   any
	}{
		{name: "header present", header: "  abc-123  ", want: "abc-123"},
		{name: "header absent", want: nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got any

			app := fiber.New()
			app.Get("/", func(fc *fiber.Ctx) error {
				processFSCTransactionID(fc)
				got = fc.UserContext().Value(models.AttrFSCTransactionID)

				return fc.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest(fiber.MethodGet, "/", http.NoBody)
			if tc.header != "" {
				req.Header.Set(models.HeaderFSCTransactionID, tc.header)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			resp.Body.Close()

			assert.Equal(t, tc.want, got)
		})
	}
}

// TestProcessTraceParent covers §3.3.1-§3.3.3: a valid traceparent header sets trace_id/span_id/
// parent_span_id on the request context (a fresh child span_id, the header's own span_id becomes the
// parent); a malformed or absent header leaves the context untouched, so applyTraceFallback still gets
// a chance to run.
func TestProcessTraceParent(t *testing.T) {
	t.Parallel()

	const (
		traceID          = "4bf92f3577b34da6a3ce929d0e0e4736"
		parentSpanID     = "00f067aa0ba902b7"
		validTraceParent = "00-" + traceID + "-" + parentSpanID + "-01"
	)

	testCases := []struct {
		name           string
		header         string
		wantTraceID    any
		wantParentSpan any
	}{
		{name: "valid header", header: validTraceParent, wantTraceID: traceID, wantParentSpan: parentSpanID},
		{name: "malformed header", header: "not-a-traceparent"},
		{name: "absent header"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotTraceID, gotSpanID, gotParentSpan any

			app := fiber.New()
			app.Get("/", func(fc *fiber.Ctx) error {
				processTraceParent(fc)
				gotTraceID = fc.UserContext().Value(models.AttrTraceID)
				gotSpanID = fc.UserContext().Value(models.AttrSpanID)
				gotParentSpan = fc.UserContext().Value(models.AttrParentSpanID)

				return fc.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest(fiber.MethodGet, "/", http.NoBody)
			if tc.header != "" {
				req.Header.Set(models.HeaderTraceParent, tc.header)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			resp.Body.Close()

			assert.Equal(t, tc.wantTraceID, gotTraceID)
			assert.Equal(t, tc.wantParentSpan, gotParentSpan)

			if tc.wantTraceID != nil {
				require.NotNil(t, gotSpanID, "the PDP's own record MUST mint a fresh child span_id, not reuse the header's")
				assert.NotEqual(t, tc.wantParentSpan, gotSpanID)
			} else {
				assert.Nil(t, gotSpanID)
			}
		})
	}
}

// TestApplyTraceFallback covers §3.3.3's precedence rule and the §4.1.1.1 Example 10 fallback: the HTTP
// header wins when present; absent that, a traceparent embedded in the AuthZEN request body's context
// object is used (and its span_id becomes the parent, exactly like the header path); absent both, a
// fresh trace/span pair is generated with no parent.
func TestApplyTraceFallback(t *testing.T) {
	t.Parallel()

	const (
		headerTraceID     = "4bf92f3577b34da6a3ce929d0e0e4736"
		headerTraceParent = "00-" + headerTraceID + "-00f067aa0ba902b7-01"
	)

	const (
		bodyTraceID     = "12345678123456781234567812345678"
		bodySpanID      = "1234567812345678"
		bodyTraceParent = "00-" + bodyTraceID + "-" + bodySpanID + "-01"
	)

	t.Run("HTTP header takes precedence over a body-embedded traceparent", func(t *testing.T) {
		t.Parallel()

		var gotTraceID any

		app := fiber.New()
		app.Get("/", func(fc *fiber.Ctx) error {
			processTraceParent(fc)
			applyTraceFallback(fc, models.NewAttributeSet(map[string]any{models.AttrTraceParent: bodyTraceParent}))
			gotTraceID = fc.UserContext().Value(models.AttrTraceID)

			return fc.SendStatus(fiber.StatusOK)
		})

		req := httptest.NewRequest(fiber.MethodGet, "/", http.NoBody)
		req.Header.Set(models.HeaderTraceParent, headerTraceParent)

		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, headerTraceID, gotTraceID)
	})

	t.Run("body-embedded traceparent used when no header is present", func(t *testing.T) {
		t.Parallel()

		var gotTraceID, gotParentSpan any

		app := fiber.New()
		app.Get("/", func(fc *fiber.Ctx) error {
			processTraceParent(fc) // no header on this request: no-op
			applyTraceFallback(fc, models.NewAttributeSet(map[string]any{models.AttrTraceParent: bodyTraceParent}))
			gotTraceID = fc.UserContext().Value(models.AttrTraceID)
			gotParentSpan = fc.UserContext().Value(models.AttrParentSpanID)

			return fc.SendStatus(fiber.StatusOK)
		})

		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", http.NoBody), -1)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, bodyTraceID, gotTraceID)
		assert.Equal(t, bodySpanID, gotParentSpan, "the body-embedded span_id MUST become the parent, same as the header path")
	})

	t.Run("fresh trace generated when neither header nor body has one", func(t *testing.T) {
		t.Parallel()

		var gotTraceID, gotParentSpan any

		app := fiber.New()
		app.Get("/", func(fc *fiber.Ctx) error {
			applyTraceFallback(fc, nil)
			gotTraceID = fc.UserContext().Value(models.AttrTraceID)
			gotParentSpan = fc.UserContext().Value(models.AttrParentSpanID)

			return fc.SendStatus(fiber.StatusOK)
		})

		resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", http.NoBody), -1)
		require.NoError(t, err)
		resp.Body.Close()

		require.NotNil(t, gotTraceID)
		assert.Len(t, gotTraceID, models.TraceIDLength)
		assert.Nil(t, gotParentSpan, "a freshly generated trace has no parent")
	})
}

// TestAuthHandler_Evaluation_BodyEmbeddedTraceParentFallback covers Logius ADL §4.1.1.1 Example 10
// end-to-end through the real Evaluation handler and a real ADL logger: with no traceparent HTTP header,
// the record persisted for the request MUST use the trace embedded in the body's context object.
func TestAuthHandler_Evaluation_BodyEmbeddedTraceParentFallback(t *testing.T) {
	t.Parallel()

	const (
		bodyTraceID     = "12345678123456781234567812345678"
		bodyTraceParent = "00-" + bodyTraceID + "-1234567812345678-01"
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exporter := tracetest.NewInMemoryExporter()
	decisionLogger, err := decisions.New(ctx, "test", otel.WithExporter(exporter), otel.WithBatchTimeout(10*time.Millisecond))
	require.NoError(t, err)

	dummy := slog.New(slog2.NewDummyHandler(slog.LevelError))
	ep := pep.New(ctx, dummy)
	ip := pip.New(ctx, dummy, pip.WithFileStore("../../../testdata/pip", true))
	ap := pap.New(ctx, dummy, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
	controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(dummy))

	auth := NewAuthHandlerZEN(dummy, adl.New(decisionLogger), controller, "http://localhost/v1")

	app := fiber.New()
	app.Post("/v1/authzen", auth.Evaluation)

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_update","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"},"context":{"traceparent":"` + bodyTraceParent + `"}}`

	req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen", bytes.NewReader([]byte(in)))
	// Deliberately no traceparent header, so the body-embedded fallback is what gets exercised.

	resp, err2 := app.Test(req, -1)
	require.NoError(t, err2)
	resp.Body.Close()
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	spans := exporter.GetSpans()

	require.NoError(t, decisionLogger.Shutdown(ctx))

	require.Len(t, spans, 1)
	assert.Equal(t, bodyTraceID, spans[0].SpanContext.TraceID().String())
}
