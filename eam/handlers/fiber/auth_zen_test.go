package fiber

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/opa"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestAuthHandler_AuthZEN1(t *testing.T) {
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"0","reasonUser":{"en":"ok"}},"decision":true}`

	t.Run("authzen handler (1)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"0","reasonUser":{"en":"ok"}},"decision":true}`

	t.Run("authzen handler (2)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("rego"), pap.WithFileStore("../../../testdata/unittest/opa", true))
		require.NotNil(t, p2)

		controller := opa.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"GET"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"context":{"id":"0","reasonUser":{"en":"not authorized"}},"decision":false}`

	t.Run("authzen handler fail (1)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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
	in := `{"subject":{"type":"","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"title":"invalid subject"}`

	t.Run("authzen handler fail (2)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	out := `{"title":"invalid action"}`

	t.Run("authzen handler fail (3)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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
	in := `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service"}}`
	out := `{"title":"invalid resource"}`

	t.Run("authzen handler fail (4)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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
