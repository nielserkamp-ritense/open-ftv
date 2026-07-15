package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
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
			wantLog: 4,
		},
		{
			name:    "OPA/Rego",
			cfg:     &config.Config{PAP: config2.PAP{Language: "opa", Store: "../../../testdata/unittest/rego"}},
			wantLog: 3,
		},
		{
			// Fail-closed (secured) mode requires OIDC; without a JWKS URL the manager
			// must refuse to initialize rather than boot and deny every request.
			name:     "fail-closed without OIDC fails",
			cfg:      &config.Config{PAP: config2.PAP{Language: "CEDAR", Store: "../../../testdata/unittest/cedar"}, Authorization: config2.Authorization{FailClosedOnEmpty: true}},
			wantFail: true,
			wantLog:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			s := &Services{ctx: context.Background(), logger: logger, cfg: tc.cfg}
			s.l = models.LanguageFromString(tc.cfg.Language)
			// Mirror the router: newAuth consumes the shared PAP, which must exist first.
			s.pap, _ = s.newPAP()

			auth := s.newAuth()
			if tc.wantFail {
				require.Nil(t, auth)
				assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
			} else {
				require.NotNil(t, auth)
				assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
				assert.NotNil(t, auth.Controller())
			}
		})
	}
}
