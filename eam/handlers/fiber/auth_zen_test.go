package fiber

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// fakeController implements pdp.Controller but only overrides Authorize; the
// handler under test invokes no other method. It returns a fixed response,
// mimicking the ODRL-geo engine that carries obligations in Response.Attributes.
type fakeController struct {
	pdp.Controller
	resp *models.Response
}

func (c *fakeController) Authorize(string, *models.PARC) (*models.Response, error) {
	return c.resp, nil
}

// TestAuthHandler_AuthZEN_Obligations verifies that ODRL obligations, the
// governing policy uid and the decision reason from Response.Attributes are
// propagated into the AuthZEN Decision context.
func TestAuthHandler_AuthZEN_Obligations(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"https://example.gov.nl/party/waterschap","id":"oorg41000"},"action":{"name":"read","properties":{"method":"GET"}},"resource":{"type":"tile","id":"https://api.example.gov.nl/bro/gmw/tiles/8/42/97"}}`

	resp := &models.Response{
		Allowed:   true,
		Message:   "permitted by urn:policy:geo-aanbod",
		PolicyKey: "urn:policy:geo-aanbod",
		Attributes: map[string]any{
			"decision_reason": "permitted by policy",
			"policy_uid":      "urn:policy:geo-aanbod",
			"obligations": []map[string]any{
				{"id": "https://standaarden.overheid.nl/odrl-geo-nl/filterFeatures", "properties": map[string]any{"area": "urn:area:defensie"}},
			},
		},
	}

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	auth := NewAuthHandlerZEN(logger, nil, &fakeController{resp: resp})
	require.NotNil(t, auth)

	app := fiber.New()
	app.Post("/v1/authzen", auth.Authorize)

	req, err := http.NewRequest(fiber.MethodPost, "/v1/authzen", bytes.NewReader([]byte(in)))
	require.NoError(t, err)

	httpResp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer httpResp.Body.Close()
	assert.Equal(t, fiber.StatusOK, httpResp.StatusCode)

	data, err := io.ReadAll(httpResp.Body)
	require.NoError(t, err)

	var out authzen.AuthorizationResponse
	require.NoError(t, json.Unmarshal(data, &out))

	assert.True(t, out.Decision)
	require.NotNil(t, out.Context)
	assert.Equal(t, "urn:policy:geo-aanbod", out.Context.Id)
	assert.Equal(t, "permitted by urn:policy:geo-aanbod", out.Context.ReasonUser["en"])
	assert.Equal(t, "urn:policy:geo-aanbod", out.Context.ReasonAdmin["policy"])
	assert.Equal(t, "permitted by policy", out.Context.ReasonAdmin["reason"])

	require.Len(t, out.Context.Obligations, 1)
	assert.Equal(t, "https://standaarden.overheid.nl/odrl-geo-nl/filterFeatures", out.Context.Obligations[0]["id"])

	// A9/B19/B20: the context also carries a flat reason string and
	// audit_identifiers.policy_version, which the BRO PEP's / mock-PDP read.
	var raw struct {
		Context struct {
			Reason           string         `json:"reason"`
			AuditIdentifiers map[string]any `json:"audit_identifiers"`
		} `json:"context"`
	}
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Equal(t, "permitted by policy", raw.Context.Reason)
	assert.Equal(t, "urn:policy:geo-aanbod", raw.Context.AuditIdentifiers["policy_version"])
}

func TestAuthHandler_AuthZEN1(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_update","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"0","reasonUser":{"en":"ok"}},"decision":true}`

	t.Run("authzen handler (1)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, ip)

		ap := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, ap)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

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
	out := `{"context":{"id":"0","reasonUser":{"en":"ok"}},"decision":true}`

	t.Run("authzen handler (2)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, ip)

		ap := pap2.New(ctx, logger, pap2.WithLanguage("rego"), pap2.WithFileStore("../../../testdata/unittest/opa", true))
		require.NotNil(t, ap)

		controller := opa_embedded.NewController(pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

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

func TestAuthHandler_AuthZEN_Fail1(t *testing.T) {
	t.Parallel()

	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"GET"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"0","reasonUser":{"en":"not authorized"}},"decision":false}`

	t.Run("authzen handler fail (1)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, ip)

		ap := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, ap)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

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
	out := `{"title":"invalid subject"}`

	t.Run("authzen handler fail (2)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

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
	out := `{"title":"invalid action"}`

	t.Run("authzen handler fail (3)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

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
	out := `{"title":"invalid resource"}`

	t.Run("authzen handler fail (4)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, ip)

		ap := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, ap)

		controller := cedar_embedded.NewController(pdp.WithPEP(ep), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerZEN(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/authzen", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/authzen", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

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
