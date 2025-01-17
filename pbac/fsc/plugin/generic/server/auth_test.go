package server

import (
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/fsc/plugin/generic/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name     string
		cfg      *config.Config
		wantFail bool
		wantLog  int
	}{
		{
			name:     "unsupported policy language",
			cfg:      &config.Config{PolicyLanguage: "ai-magic", PolicyStore: "../../../../../testdata/unittest/ai"},
			wantFail: true,
			wantLog:  1,
		},
		{
			name:    "Cedar",
			cfg:     &config.Config{PolicyLanguage: "CEDAR", PolicyStore: "../../../../../testdata/unittest/cedar"},
			wantLog: 4,
		},
		{
			name:    "Cerbos",
			cfg:     &config.Config{PolicyLanguage: "Cerbos", PolicyStore: "../../../../../testdata/unittest/cerbos"},
			wantLog: 4,
		},
		{
			name:    "OpenFGA",
			cfg:     &config.Config{PolicyLanguage: "OpenFGA", PolicyStore: "../../../../../testdata/unittest/openfga"},
			wantLog: 6,
		},
		{
			name:    "OPA/Rego",
			cfg:     &config.Config{PolicyLanguage: "opa", PolicyStore: "../../../../../testdata/unittest/rego"},
			wantLog: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			auth := New(nil, tc.cfg, logger, nil)
			if tc.wantFail {
				require.Nil(t, auth)
				assert.Equal(t, tc.wantLog, h.Count())
			} else {
				require.NotNil(t, auth)
				assert.GreaterOrEqual(t, tc.wantLog, h.Count())
				assert.NotNil(t, auth.Controller())

				srv := fiber.New()
				srv.Get("/fsc", auth.AuthFSC)
				srv.Get("/zen", auth.AuthZEN)

				req := httptest.NewRequest("GET", "/fsc", nil)
				resp, err := srv.Test(req, 1)
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

				req = httptest.NewRequest("GET", "/zen", nil)
				resp, err = srv.Test(req, 1)
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
			}
		})
	}
}
