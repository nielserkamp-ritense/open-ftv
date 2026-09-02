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
	}{
		{
			name: "Cedar",
			cfg:  &config.Config{PAP: config2.PAP{Language: "CEDAR"}},
		},
		{
			// Fail-closed (secured) mode requires OIDC; without a JWKS URL the manager
			// must refuse to initialize rather than boot and deny every request.
			name:     "fail-closed without OIDC fails",
			cfg:      &config.Config{PAP: config2.PAP{Language: "CEDAR"}, Authorization: config2.Authorization{FailClosedOnEmpty: true}},
			wantFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

			s := &Services{ctx: context.Background(), logger: logger, cfg: tc.cfg}
			s.l = models.LanguageFromString(tc.cfg.Language)

			ap, err := s.cfg.NewSelfAuthzPAP(s.ctx, s.logger)
			require.NoError(t, err)

			s.pap = ap

			auth := s.newAuth()
			if tc.wantFail {
				require.Nil(t, auth)

				return
			}

			require.NotNil(t, auth)
			require.NotNil(t, auth.Controller())
			assert.Same(t, ap, auth.Controller().GetPAP())
		})
	}
}
