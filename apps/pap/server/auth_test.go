package server

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/pap/config"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
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
			cfg:      &config.Config{PolicyLanguage: "ai-magic", PolicyStore: "../../../testdata/unittest/ai"},
			wantFail: true,
			wantLog:  3,
		},
		{
			name:    "Cedar",
			cfg:     &config.Config{PolicyLanguage: "CEDAR", PolicyStore: "../../../testdata/unittest/cedar"},
			wantLog: 4,
		},
		{
			name:    "Cerbos",
			cfg:     &config.Config{PolicyLanguage: "Cerbos", PolicyStore: "../../../testdata/unittest/cerbos"},
			wantLog: 4,
		},
		{
			name:    "OpenFGA",
			cfg:     &config.Config{PolicyLanguage: "OpenFGA", PolicyStore: "../../../testdata/unittest/openfga"},
			wantLog: 5,
		},
		{
			name:    "OPA/Rego",
			cfg:     &config.Config{PolicyLanguage: "opa", PolicyStore: "../../../testdata/unittest/rego"},
			wantLog: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			s := &service{ctx: context.Background(), logger: logger, cfg: tc.cfg}
			s.l = models.LanguageFromString(tc.cfg.PolicyLanguage)

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
