package fiber

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestFormatRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		path       string
		method     string
		headers    map[string][]string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid path",
			path:       "/v1/toast",
			method:     "GET",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "GET",
			path:       "/v1/test",
			method:     "GET",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST",
			path:       "/v1/test",
			method:     "POST",
			headers:    map[string][]string{"Content-Type": {"application/json"}},
			body:       `{"hello":"world"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "DELETE",
			path:       "/v1/test",
			method:     "DELETE",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p1 := pip2.New(ctx, logger, pip2.WithFileStore("../../../testdata/pip", true))
			require.NotNil(t, p1)

			p2 := pap2.New(ctx, logger, pap2.WithLanguage("rego"), pap2.WithFileStore("../../../testdata/policies/opa", true))
			require.NotNil(t, p2)

			controller := opa_embedded.NewController(pdp.WithContext(ctx), pdp.WithPIP(p1), pdp.WithPAP(p2), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			var a *authorization.Request
			f := func(req *fiber.Ctx) error {
				a = FormatRequest(req)
				return req.SendStatus(fiber.StatusOK)
			}

			srv := fiber.New()
			srv.Options("/v1/test", f)
			srv.Head("/v1/test", f)
			srv.Get("/v1/test", f)
			srv.Post("/v1/test", f)
			srv.Put("/v1/test", f)
			srv.Delete("/v1/test", f)

			req := httptest.NewRequestWithContext(ctx, tc.method, tc.path, strings.NewReader(tc.body))

			for key := range tc.headers {
				list := tc.headers[key]
				for i := range list {
					req.Header.Add(key, list[i])
				}
			}

			resp, err2 := srv.Test(req)

			require.NoError(t, err2)
			require.NotNil(t, resp)
			defer resp.Body.Close()

			require.Equal(t, tc.wantStatus, resp.StatusCode)
			if tc.wantStatus == http.StatusOK {
				require.NotNil(t, a)

				assert.NotNil(t, a.UID)
				assert.NotNil(t, a.URL)
				assert.Equal(t, tc.path, a.URL.Path)
				assert.Equal(t, tc.method, a.Method)
				assert.Equal(t, tc.body, string(a.Body))

				for key := range tc.headers {
					h1, h2 := tc.headers[key], a.Headers[key]
					assert.EqualValues(t, h1, h2)
				}
			}
		})
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		resp       *models.Response
		err        error
		wantOK     bool
		wantErr    bool
		wantHeader bool
	}{
		{
			name:       "authentication error",
			err:        &authentication.ErrUnauthenticated{},
			wantHeader: true,
		},
		{
			name: "other error",
			err:  fmt.Errorf("oh my"),
		},
		{
			name: "nil response",
		},
		{
			name: "not allowed",
			resp: &models.Response{},
		},
		{
			name:   "allowed",
			resp:   &models.Response{Allowed: true},
			wantOK: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			log := slog.New(h)

			var got bool

			srv := fiber.New()
			srv.Get("/v1/attributes", func(req *fiber.Ctx) error {
				_, err := Check(req, tc.resp, identity.NewUnknownPrincipal(), tc.err, log)
				got = err == nil

				return err
			})

			req := httptest.NewRequest(fiber.MethodGet, "/v1/attributes", nil)
			resp, err2 := srv.Test(req)

			if tc.wantErr {
				require.Error(t, err2)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err2)
				assert.NotNil(t, resp)
				defer resp.Body.Close()
				assert.Equal(t, tc.wantOK, got)

				if tc.wantHeader {
					h := resp.Header.Get(fiber.HeaderWWWAuthenticate)
					assert.NotEmpty(t, h)
					assert.True(t, strings.HasPrefix(h, "Basic realm"))
				}
			}
		})
	}
}
