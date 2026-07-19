package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
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
			wantLog: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			s := &Services{ctx: context.Background(), cfg: tc.cfg, logger: logger, l: models.LanguageFromString(tc.cfg.Language)}

			auth := s.newAuth("")
			if tc.wantFail {
				require.Nil(t, auth)
				assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
			} else {
				require.NotNil(t, auth)
				assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
				assert.NotNil(t, auth.controller)

				srv := fiber.New()
				srv.Get("/zen", auth.zen.Evaluation)

				req := httptest.NewRequest("GET", "/zen", http.NoBody)
				resp, err := srv.Test(req)
				require.NoError(t, err)
				require.NotNil(t, resp)
				defer resp.Body.Close()
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
			}
		})
	}
}
