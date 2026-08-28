package fiber

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNewPoliciesHandler(t *testing.T) {
	t.Parallel()

	t.Run("new policies handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
		require.NoError(t, err)
		require.NotNil(t, p1)

		p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
		require.NoError(t, err)
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		ph := NewPoliciesHandler(logger, controller.GetPAP(), nil)
		require.NotNil(t, ph)
	})
}

func TestPoliciesHandler_GetPolicies(t *testing.T) {
	t.Parallel()

	t.Run("get policies", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
		require.NoError(t, err)
		require.NotNil(t, p1)

		p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
		require.NoError(t, err)
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ph := NewPoliciesHandler(logger, controller.GetPAP(), auth)
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/policies", ph.GetPolicies)

		req := httptest.NewRequest(fiber.MethodGet, "/v1/policies", nil)
		resp, err2 := srv.Test(req)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, PoliciesVersion, resp.Header.Get(HeaderVersion))
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list []*policies.Policy

		err = json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 6)
	})
}

func TestPoliciesHandler_GetPolicies_NotFound(t *testing.T) {
	t.Parallel()

	t.Run("get policies - not found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/non_existing_folder", true))
		require.NoError(t, err)
		require.NotNil(t, p1)

		p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/non_existing_folder", true))
		require.NoError(t, err)
		require.NotNil(t, p2)

		controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
		require.NotNil(t, controller)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ph := NewPoliciesHandler(logger, controller.GetPAP(), auth)
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/policies", ph.GetPolicies)

		req := httptest.NewRequest(fiber.MethodGet, "/v1/policies", nil)
		resp, err2 := srv.Test(req)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, PoliciesVersion, resp.Header.Get(HeaderVersion))
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestPoliciesHandler_GetPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		language   string
		id         string
		wantStatus int
		wantVer    string
	}{
		{name: "no ID", language: "cedar", wantStatus: fiber.StatusNotFound},
		{name: "very long ID", language: "cedar", id: strings.Repeat("x", 501), wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "bad ID", language: "cedar", id: "xyz", wantStatus: fiber.StatusNotFound, wantVer: PoliciesVersion},
		{name: "good ID", language: "cedar", id: "subsidies.cedar", wantStatus: fiber.StatusOK, wantVer: PoliciesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)
			require.NotNil(t, p1)

			p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
			require.NoError(t, err)
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
			require.NotNil(t, auth)

			ph := NewPoliciesHandler(logger, controller.GetPAP(), auth)
			require.NotNil(t, ph)

			srv := fiber.New()
			srv.Get("/v1/policy/:id", ph.GetPolicy)

			req := httptest.NewRequest(fiber.MethodGet, fmt.Sprintf("/v1/policy/%s", tc.id), nil)
			resp, err2 := srv.Test(req)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
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
	t.Parallel()

	badURL := `{
 "language":"cedar",
 "metadata": {
  "title": "titel",
  "rvvaID": "id1",
  "url": "http://localhost:29171/policy/xyzqqq"
 }
}`

	goodURL := `{
 "language":"cedar",
 "metadata": {
  "title": "titel2",
  "rvvaID": "id1",
  "url": "https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/-/blob/f086f33b2e2fc51e494e0a22bc7bbcc045847250/testdata/policies/cedar/brp/subsidies.cedar"
 }
}`

	testCases := []struct {
		name       string
		language   string
		id         string
		body       io.Reader
		timeout    time.Duration
		wantStatus int
		wantVer    string
	}{
		{name: "no ID", language: "cedar", body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusNotFound},
		{name: "very long ID", language: "cedar", id: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "no body", language: "cedar", id: "xyz", timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "bad body", language: "cedar", id: "xyz", body: bytes.NewBufferString("my policy 1.0"), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "bad url", language: "cedar", id: "xyz", body: bytes.NewBufferString(badURL), timeout: 15 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "duplicate id", language: "cedar", id: "subsidies.cedar", body: bytes.NewBufferString(goodURL), timeout: 15 * time.Second, wantStatus: fiber.StatusConflict, wantVer: PoliciesVersion},
		{name: "all good", language: "cedar", id: "xyz", body: bytes.NewBufferString(goodURL), timeout: 15 * time.Second, wantStatus: fiber.StatusCreated, wantVer: PoliciesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)
			require.NotNil(t, p1)

			p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
			require.NoError(t, err)
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
			require.NotNil(t, auth)

			ph := NewPoliciesHandler(logger, controller.GetPAP(), auth)
			require.NotNil(t, ph)

			srv := fiber.New()
			srv.Post("/v1/policy/:id", ph.PostPolicy)

			req := httptest.NewRequest(fiber.MethodPost, fmt.Sprintf("/v1/policy/%s", tc.id), tc.body)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, int(tc.timeout/time.Millisecond))

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
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
	t.Parallel()

	badURL := `{
 "language":"cedar",
 "metadata": {
  "title": "titel",
  "rvvaID": "id1",
  "url": "http://bad.url.xyz:\000/policy/xyzqqq"
 }
}`

	goodURL := `{
 "language":"cedar",
 "metadata": {
  "title": "titel2",
  "url": "https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/-/blob/f086f33b2e2fc51e494e0a22bc7bbcc045847250/testdata/policies/cedar/brp/subsidies.cedar"
 }
}`

	testCases := []struct {
		name       string
		language   string
		id         string
		body       io.Reader
		timeout    time.Duration
		wantStatus int
		wantVer    string
	}{
		{name: "no ID", language: "cedar", body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusNotFound},
		{name: "very long ID", language: "cedar", id: strings.Repeat("x", 501), body: bytes.NewBufferString(""), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "no body", language: "cedar", id: "xyz", timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "bad body", language: "cedar", id: "xyz", body: bytes.NewBufferString("my policy 1.0"), timeout: 1 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "bad url", language: "cedar", id: "xyz", body: bytes.NewBufferString(badURL), timeout: 15 * time.Second, wantStatus: fiber.StatusBadRequest, wantVer: PoliciesVersion},
		{name: "not found", language: "cedar", id: "xyz", body: bytes.NewBufferString(goodURL), timeout: 15 * time.Second, wantStatus: fiber.StatusNotFound, wantVer: PoliciesVersion},
		{name: "all good", language: "cedar", id: "subsidies.cedar", body: bytes.NewBufferString(goodURL), timeout: 15 * time.Second, wantStatus: fiber.StatusOK, wantVer: PoliciesVersion},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)
			require.NotNil(t, p1)

			p2, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore("../../../testdata/policies/cedar", true))
			require.NoError(t, err)
			require.NotNil(t, p2)

			controller := cedar_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
			require.NotNil(t, auth)

			ph := NewPoliciesHandler(logger, controller.GetPAP(), auth)
			require.NotNil(t, ph)

			srv := fiber.New()
			srv.Put("/v1/policy/:id", ph.PutPolicy)

			req := httptest.NewRequest(fiber.MethodPut, fmt.Sprintf("/v1/policy/%s", tc.id), tc.body)
			req.Header.Add(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			resp, err2 := srv.Test(req, int(tc.timeout/time.Millisecond))

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			assert.Equal(t, tc.wantVer, resp.Header.Get(HeaderVersion))
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

func TestPoliciesHandler_buildPolicy(t *testing.T) {
	t.Parallel()

	const cedar = "permit(principal, action, resource);"

	t.Run("neither url nor data", func(t *testing.T) {
		t.Parallel()

		h := &policiesHandler{logger: slog.New(slog2.NewDummyHandler(slog.LevelDebug))}

		pol, err := h.buildPolicy(&policies.Policy{Language: "cedar"})

		require.ErrorIs(t, err, errPolUrlContent)
		assert.Nil(t, pol)
	})

	t.Run("inline data", func(t *testing.T) {
		t.Parallel()

		h := &policiesHandler{logger: slog.New(slog2.NewDummyHandler(slog.LevelDebug))}

		pol, err := h.buildPolicy(&policies.Policy{Language: "cedar", Data: cedar})

		require.NoError(t, err)
		assert.NotNil(t, pol)
	})

	t.Run("url", func(t *testing.T) {
		t.Parallel()

		h := &policiesHandler{logger: slog.New(slog2.NewDummyHandler(slog.LevelDebug))}

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(cedar))
		}))
		defer srv.Close()

		p := &policies.Policy{Language: "cedar"}
		p.Metadata.Url = srv.URL

		pol, err := h.buildPolicy(p)

		require.NoError(t, err)
		assert.NotNil(t, pol)
	})

	t.Run("unreachable url", func(t *testing.T) {
		t.Parallel()

		h := &policiesHandler{logger: slog.New(slog2.NewDummyHandler(slog.LevelDebug))}

		// A server that is started and immediately stopped hands us an address that is
		// guaranteed to refuse connections.
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		srv.Close()

		p := &policies.Policy{Language: "cedar"}
		p.Metadata.Url = srv.URL

		pol, err := h.buildPolicy(p)

		require.Error(t, err)
		assert.Nil(t, pol)
	})
}
