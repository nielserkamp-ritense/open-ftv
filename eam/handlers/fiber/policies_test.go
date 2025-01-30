package fiber

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp/cedar"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewPoliciesHandler(t *testing.T) {
	t.Run("new policies handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ph := NewPoliciesHandler(logger, controller.PAP())
		require.NotNil(t, ph)
	})
}

func TestPoliciesHandler_GetPolicies(t *testing.T) {
	t.Run("get attributes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ph := NewPoliciesHandler(logger, controller.PAP())
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/policies", ph.GetPolicies)

		req := httptest.NewRequest(fiber.MethodGet, "/v1/policies", nil)
		resp, err2 := srv.Test(req, 1)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list []*policies.Policy
		err := json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.Equal(t, 4, len(list))
	})
}

func TestPoliciesHandler_GetPolicies_NotFOund(t *testing.T) {
	t.Run("get attributes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/non_existing_folder", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
		require.NotNil(t, p)

		controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/non_existing_folder", true), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ph := NewPoliciesHandler(logger, controller.PAP())
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/policies", ph.GetPolicies)

		req := httptest.NewRequest(fiber.MethodGet, "/v1/policies", nil)
		resp, err2 := srv.Test(req, 1)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

func TestPoliciesHandler_GetPolicy(t *testing.T) {
	testCases := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{name: "no ID", wantStatus: fiber.StatusNotFound},
		{name: "very long ID", id: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest},
		{name: "bad ID", id: "xyz", wantStatus: fiber.StatusNotFound},
		{name: "good ID", id: "subsidies.cedar", wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
			require.NotNil(t, p)

			controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ph := NewPoliciesHandler(logger, controller.PAP())
			require.NotNil(t, ph)

			srv := fiber.New()
			srv.Get("/v1/policy/:id", ph.GetPolicy)

			req := httptest.NewRequest(fiber.MethodGet, "/v1/policy/"+tc.id, nil)
			resp, err2 := srv.Test(req, 1)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var pol policies.Policy
				err := json.Unmarshal(b, &pol)
				require.NoError(t, err)
				assert.NotNil(t, pol)
			}
		})
	}
}

func TestPoliciesHandler_PutPolicy(t *testing.T) {
	badURL := `{
 "source": "s1",
 "target": "t1",
 "rvvaID": "id1",
 "url": "http://localhost:29171/policy/xyzqqq"
}`

	goodURL := `{
 "source": "s1",
 "target": "t1",
 "rvvaID": "id1",
 "url": "https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/-/blob/f086f33b2e2fc51e494e0a22bc7bbcc045847250/testdata/policies/cedar/brp/subsidies.cedar"
}`

	testCases := []struct {
		name       string
		id         string
		body       io.Reader
		timeout    time.Duration
		wantStatus int
	}{
		{name: "no ID", body: bytes.NewBufferString(""), timeout: time.Millisecond, wantStatus: fiber.StatusNotFound},
		{name: "very long ID", id: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "no body", id: "xyz", timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "bad body", id: "xyz", body: bytes.NewBufferString("my policy 1.0"), timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "bad url", id: "xyz", body: bytes.NewBufferString(badURL), timeout: 5 * time.Second, wantStatus: fiber.StatusBadRequest},
		{name: "duplicate id", id: "subsidies.cedar", body: bytes.NewBufferString(goodURL), timeout: 5 * time.Second, wantStatus: fiber.StatusConflict},
		{name: "all good", id: "xyz", body: bytes.NewBufferString(goodURL), timeout: 5 * time.Second, wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
			require.NotNil(t, p)

			controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ph := NewPoliciesHandler(logger, controller.PAP())
			require.NotNil(t, ph)

			srv := fiber.New()
			srv.Put("/v1/policy/:id", ph.PutPolicy)

			req := httptest.NewRequest(fiber.MethodPut, "/v1/policy/"+tc.id, tc.body)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, int(tc.timeout/time.Millisecond))

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var pol policies.Policy
				err := json.Unmarshal(b, &pol)
				require.NoError(t, err)
				assert.NotNil(t, pol)
			}
		})
	}
}

func TestPoliciesHandler_PostPolicy(t *testing.T) {
	badURL := `{
 "source": "s1",
 "target": "t1",
 "rvvaID": "id1",
 "url": "http://bad.url.xyz:\000/policy/xyzqqq"
}`

	goodURL := `{
 "source": "s1",
 "target": "t1",
 "rvvaID": "id1",
 "url": "https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/-/blob/f086f33b2e2fc51e494e0a22bc7bbcc045847250/testdata/policies/cedar/brp/subsidies.cedar"
}`

	testCases := []struct {
		name       string
		id         string
		body       io.Reader
		timeout    time.Duration
		wantStatus int
	}{
		{name: "no ID", body: bytes.NewBufferString(""), timeout: time.Millisecond, wantStatus: fiber.StatusNotFound},
		{name: "very long ID", id: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "no body", id: "xyz", timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "bad body", id: "xyz", body: bytes.NewBufferString("my policy 1.0"), timeout: time.Millisecond, wantStatus: fiber.StatusBadRequest},
		{name: "bad url", id: "xyz", body: bytes.NewBufferString(badURL), timeout: 5 * time.Second, wantStatus: fiber.StatusBadRequest},
		{name: "not found", id: "xyz", body: bytes.NewBufferString(goodURL), timeout: 5 * time.Second, wantStatus: fiber.StatusNotFound},
		{name: "all good", id: "subsidies.cedar", body: bytes.NewBufferString(goodURL), timeout: 5 * time.Second, wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
			require.NotNil(t, p)

			controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ph := NewPoliciesHandler(logger, controller.PAP())
			require.NotNil(t, ph)

			srv := fiber.New()
			srv.Post("/v1/policy/:id", ph.PostPolicy)

			req := httptest.NewRequest(fiber.MethodPost, "/v1/policy/"+tc.id, tc.body)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, int(tc.timeout/time.Millisecond))

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var pol policies.Policy
				err := json.Unmarshal(b, &pol)
				require.NoError(t, err)
				assert.NotNil(t, pol)
			}
		})
	}
}

func TestPoliciesHandler_DeletePolicy(t *testing.T) {
	testCases := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{name: "no ID", wantStatus: fiber.StatusNotFound},
		{name: "very long ID", id: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest},
		{name: "not found", id: "xyz", wantStatus: fiber.StatusNotFound},
		{name: "all good", id: "subsidies.cedar", wantStatus: fiber.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{Ctx: ctx, Store: "../../../testdata/pip", Recurse: true, Logger: logger, NewAttributes: cedar.NewAttributeBuilder(logger), NewEntities: cedar.NewEntityBuilder(logger)})
			require.NotNil(t, p)

			controller := cedar.NewController(pdp.WithContext(ctx), pdp.WithPIP(p), pdp.WithStore("../../../testdata/policies/cedar", true), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			ph := NewPoliciesHandler(logger, controller.PAP())
			require.NotNil(t, ph)

			srv := fiber.New()
			srv.Delete("/v1/policy/:id", ph.DeletePolicy)

			req := httptest.NewRequest(fiber.MethodDelete, "/v1/policy/"+tc.id, nil)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, 1)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			if tc.wantStatus == fiber.StatusOK {
				b, err3 := io.ReadAll(resp.Body)
				require.NoError(t, err3)
				require.NotNil(t, b)

				var pol policies.Policy
				err := json.Unmarshal(b, &pol)
				require.NoError(t, err)
				assert.NotNil(t, pol)
			}
		})
	}
}
