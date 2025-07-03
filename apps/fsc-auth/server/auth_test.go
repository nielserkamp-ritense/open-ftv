package server

import (
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/fsc-auth/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		cfg      *config.Config
		wantFail bool
		wantLog  int
	}{
		{
			name:     "unsupported policy language",
			cfg:      &config.Config{PAP: config2.PAP{Language: "ai-magic", Store: "../../../testdata/unittest/ai"}},
			wantFail: true,
			wantLog:  3,
		},
		{
			name:    "Cedar",
			cfg:     &config.Config{PAP: config2.PAP{Language: "CEDAR", Store: "../../../testdata/unittest/cedar"}},
			wantLog: 4,
		},
		{
			name:    "Cerbos",
			cfg:     &config.Config{PAP: config2.PAP{Language: "Cerbos", Store: "../../../testdata/unittest/cerbos"}},
			wantLog: 4,
		},
		{
			name:    "OpenFGA",
			cfg:     &config.Config{PAP: config2.PAP{Language: "OpenFGA", Store: "../../../testdata/unittest/openfga"}},
			wantLog: 5,
		},
		{
			name:    "OPA/Rego",
			cfg:     &config.Config{PAP: config2.PAP{Language: "opa", Store: "../../../testdata/unittest/rego"}},
			wantLog: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			auth := New(nil, tc.cfg, logger)
			if tc.wantFail {
				require.Nil(t, auth)
				assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
			} else {
				require.NotNil(t, auth)
				assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
				assert.NotNil(t, auth.Controller())

				srv := fiber.New()
				srv.Get("/fsc", auth.AuthFSC)
				srv.Get("/zen", auth.AuthZEN)

				req := httptest.NewRequest("GET", "/fsc", nil)
				resp, err := srv.Test(req, 100)
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

				req = httptest.NewRequest("GET", "/zen", nil)
				resp, err = srv.Test(req, 100)
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
			}
		})
	}
}
