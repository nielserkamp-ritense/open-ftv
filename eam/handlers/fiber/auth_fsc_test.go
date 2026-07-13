package fiber

// **Deprecation warning**
// The FSC auth-plugin is deprecated, as it is no longer in use.
// THis code will be removed in a future version.

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestAuthHandler_FSC(t *testing.T) {
	t.Parallel()

	in := `{"input":{"method":"POST","path":"/x/y"}}`
	out := `{"result":{"allowed":true,"status":{"reason":"ok"}}}`

	t.Run("fsc handler", func(t *testing.T) {
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

		auth := NewAuthHandlerFSC(logger, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/auth", buf)
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
		assert.GreaterOrEqual(t, 9, h.Count())
	})
}

func TestAuthHandler_FSC_Fail1(t *testing.T) {
	t.Parallel()

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

		controller := opa_embedded.NewController(pdp.WithPEP(ep), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := NewAuthHandlerFSC(logger, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/auth", buf)
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
		assert.GreaterOrEqual(t, h.Count(), 11)
	})
}

func TestAuthHandler_FSC_Fail2(t *testing.T) {
	t.Parallel()

	in := `{"input":{}}`
	out := `{"title":"invalid method"}`

	t.Run("fsc handler fail (2)", func(t *testing.T) {
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

		auth := NewAuthHandlerFSC(logger, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/auth", buf)
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
		assert.GreaterOrEqual(t, h.Count(), 6)
	})
}

func TestAuthHandler_FSC_Fail3(t *testing.T) {
	t.Parallel()

	in := `[]`
	out := `{"title":"invalid input data"}`

	t.Run("fsc handler fail (3)", func(t *testing.T) {
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

		auth := NewAuthHandlerFSC(logger, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/auth", buf)
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
		assert.GreaterOrEqual(t, h.Count(), 6)
	})
}

func TestAuthHandler_FSC_Fail4(t *testing.T) {
	t.Parallel()

	in := `{"input":{"method":"POST"}}`
	out := `{"title":"invalid path"}`

	t.Run("fsc handler fail (4)", func(t *testing.T) {
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

		auth := NewAuthHandlerFSC(logger, controller)
		require.NotNil(t, auth)

		app := fiber.New()
		app.Post("/v1/auth", auth.Evaluation)

		buf := bytes.NewReader([]byte(in))

		req := httptest.NewRequest(fiber.MethodPost, "/v1/auth", buf)
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
		assert.GreaterOrEqual(t, h.Count(), 6)
	})
}
