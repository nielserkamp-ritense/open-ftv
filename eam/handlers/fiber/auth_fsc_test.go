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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestAuthHandler_FSC(t *testing.T) {
	in := `{"input":{"method":"POST","path":"/x/y"}}`
	out := `{"result":{"allowed":true,"status":{"reason":"ok"}}}`

	t.Run("fsc handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true), pip.WithFactories(cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)))
		require.NotNil(t, ip)

		ap := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, ap)

		controller := cedar.NewController(pdp.WithPEP(ep), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerFSC(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, AuthFSCVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 8, h.Count())
	})
}

func TestAuthHandler_FSC_Fail1(t *testing.T) {
	in := `{"input":{"method":"GET","path":"/x/y","headers":{"doelbinding":["subsidies"]}}}`
	out := `{"result":{"allowed":false,"status":{"reason":"not authorized"}}}`

	t.Run("fsc handler fail (1)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		ep := pep.New(ctx, logger)

		ip := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, ip)

		ap := pap.New(ctx, logger, pap.WithLanguage("rego"), pap.WithFileStore("../../../testdata/unittest/opa", true))
		require.NotNil(t, ap)

		controller := opa.NewController(pdp.WithPEP(ep), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerFSC(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, AuthFSCVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 10, h.Count())
	})
}

func TestAuthHandler_FSC_Fail2(t *testing.T) {
	in := `{"input":{}}`
	out := `{"title":"invalid method"}`

	t.Run("fsc handler fail (2)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true), pip.WithFactories(cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerFSC(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, AuthFSCVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 6, h.Count())
	})
}

func TestAuthHandler_FSC_Fail3(t *testing.T) {
	in := `[]`
	out := `{"title":"invalid input data"}`

	t.Run("fsc handler fail (3)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true), pip.WithFactories(cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerFSC(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, AuthFSCVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 6, h.Count())
	})
}

func TestAuthHandler_FSC_Fail4(t *testing.T) {
	in := `{"input":{"method":"POST"}}`
	out := `{"title":"invalid path"}`

	t.Run("fsc handler fail (4)", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1 := pip.New(ctx, logger, pip.WithFileStore("../../../testdata/pip", true), pip.WithFactories(cedar.NewAttributeBuilder(logger), cedar.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap.New(ctx, logger, pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerFSC(logger, nil, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Authorize)

		buf := bytes.NewReader([]byte(in))

		req, err2 := http.NewRequest(fiber.MethodPost, "/v1/auth", buf)
		require.NoError(t, err2)
		require.NotNil(t, req)

		resp, err3 := app.Test(req, -1)
		require.NoError(t, err3)
		require.NotNil(t, resp)

		assert.Equal(t, AuthFSCVersion, resp.Header.Get(HeaderVersion))
		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		defer resp.Body.Close()

		data, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, data)

		assert.Equal(t, out, string(data))
		assert.GreaterOrEqual(t, 6, h.Count())
	})
}
