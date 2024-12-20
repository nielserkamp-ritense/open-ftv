package handlers

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gjuyn/go-config/config"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
	cfg2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewPoliciesHandler(t *testing.T) {
	t.Run("test new policies handler", func(t *testing.T) {
		cfg, _, err := cfg2.New(config.NoFlags())
		require.NoError(t, err)
		require.NotNil(t, cfg)

		h := slog2.NewDummyHandler(slog.LevelInfo)

		ph := NewPoliciesHandler(cfg, slog.New(h), nil)
		require.NotNil(t, ph)
	})
}

func TestPoliciesHandler_GetPolicies(t *testing.T) {
	t.Run("test new policies handler", func(t *testing.T) {
		cfg := &cfg2.Config{
			PolicyLanguage:     "cedar",
			PolicyStore:        "../../../../../testdata/policies/cedar",
			PolicyStoreRecurse: true,
			PipStore:           "../../../../../testdata/pip",
			PipStoreRecurse:    true,
		}

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		c, err := newController(cfg, logger, nil)
		require.NoError(t, err)
		require.NotNil(t, c)

		ph := NewPoliciesHandler(cfg, logger, c)
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/policies", ph.GetPolicies)

		req := httptest.NewRequest("GET", "/v1/policies", nil)
		resp, err2 := srv.Test(req, 1)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list []*policies.Policy
		err = json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.Equal(t, 4, len(list))
	})
}

func TestPoliciesHandler_GetPolicy(t *testing.T) {
	testCases := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "no ID",
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:       "very long ID",
			id:         strings.Repeat("x", 501),
			wantStatus: fiber.StatusBadRequest,
		},
		{
			name:       "bad ID",
			id:         "xyz",
			wantStatus: fiber.StatusNotFound,
		},
		{
			name:       "good ID",
			id:         "subsidies.cedar",
			wantStatus: fiber.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &cfg2.Config{
				PolicyLanguage:     "cedar",
				PolicyStore:        "../../../../../testdata/policies/cedar",
				PolicyStoreRecurse: true,
				PipStore:           "../../../../../testdata/pip",
				PipStoreRecurse:    true,
			}

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			c, err := newController(cfg, logger, nil)
			require.NoError(t, err)
			require.NotNil(t, c)

			ph := NewPoliciesHandler(cfg, logger, c)
			require.NotNil(t, ph)

			srv := fiber.New()
			srv.Get("/v1/policy/:id", ph.GetPolicy)

			req := httptest.NewRequest("GET", "/v1/policy/"+tc.id, nil)
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
				err = json.Unmarshal(b, &pol)
				require.NoError(t, err)
				assert.NotNil(t, pol)
			}
		})
	}
}
