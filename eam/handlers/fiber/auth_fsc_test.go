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

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/authlog"
	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
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

		ip := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, ip)

		ap := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, ap)

		controller := cedar_embedded.NewController(pdp.WithPEP(ep), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
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

// captureLogger records the AuthRecords passed to it.
type captureLogger struct{ recs []*authlog.AuthRecord }

func (c *captureLogger) Log(_ context.Context, _ bool, rec *authlog.AuthRecord) error {
	c.recs = append(c.recs, rec)
	return nil
}

// F18: the FSC handler extracts the Fsc-Transaction-Id header onto the AuthRecord (which
// build.go then maps to adl.fsc.transaction_id). A request without the header leaves it empty.
func TestAuthHandler_FSC_TransactionID(t *testing.T) {
	t.Parallel()

	run := func(t *testing.T, headers string) *authlog.AuthRecord {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
		ep := pep.New(ctx, logger)
		ip := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		ap := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		controller := cedar_embedded.NewController(pdp.WithPEP(ep), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))

		cap := &captureLogger{}
		auth := NewAuthHandlerFSC(logger, cap, controller)
		app := fiber.New()
		app.Post("/v1/auth", auth.Authorize)

		in := `{"input":{"method":"POST","path":"/x/y"` + headers + `}}`
		req, err := http.NewRequest(fiber.MethodPost, "/v1/auth", bytes.NewReader([]byte(in)))
		require.NoError(t, err)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Len(t, cap.recs, 1)
		return cap.recs[0]
	}

	t.Run("header present", func(t *testing.T) {
		rec := run(t, `,"headers":{"Fsc-Transaction-Id":["0190f2c0-7a11-7000-8000-abcabcabcabc"]}`)
		assert.Equal(t, "0190f2c0-7a11-7000-8000-abcabcabcabc", rec.FSCTransactionID)
	})

	t.Run("header absent", func(t *testing.T) {
		rec := run(t, "")
		assert.Empty(t, rec.FSCTransactionID)
	})

	t.Run("case-insensitive header", func(t *testing.T) {
		rec := run(t, `,"headers":{"fsc-transaction-id":["tx-lower"]}`)
		assert.Equal(t, "tx-lower", rec.FSCTransactionID)
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

		ip := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true))
		require.NotNil(t, ip)

		ap := pap2.New(ctx, logger, pap2.WithLanguage("rego"), pap2.WithFileStore("../../../testdata/unittest/opa", true))
		require.NotNil(t, ap)

		controller := opa_embedded.NewController(pdp.WithPEP(ep), pdp.WithContext(ctx), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
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

		p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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

		p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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

		p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true), pip2.WithFactories(cedar_embedded.NewAttributeBuilder(logger), cedar_embedded.NewEntityBuilder(logger)))
		require.NotNil(t, p1)

		p2 := pap2.New(ctx, logger, pap2.WithLanguage("cedar"), pap2.WithFileStore("../../../testdata/unittest/cedar", true))
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
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
		assert.GreaterOrEqual(t, h.Count(), 6)
	})
}
