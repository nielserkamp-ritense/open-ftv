package handlers

import (
	"log/slog"
	"testing"

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
			name:    "cedar",
			cfg:     &config.Config{PolicyLanguage: "CEDAR", PolicyStore: "../../../../../testdata/unittest/cedar"},
			wantLog: 4,
		},
		{
			name:    "cerbos",
			cfg:     &config.Config{PolicyLanguage: "Cerbos", PolicyStore: "../../../../../testdata/unittest/cerbos"},
			wantLog: 3,
		},
		{
			name:    "opa",
			cfg:     &config.Config{PolicyLanguage: "opa", PolicyStore: "../../../../../testdata/unittest/rego"},
			wantLog: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			auth := New(tc.cfg, logger, nil)
			if tc.wantFail {
				require.Nil(t, auth)
				assert.Equal(t, tc.wantLog, h.Count())
			} else {
				require.NotNil(t, auth)
				assert.Equal(t, tc.wantLog, h.Count())
			}
		})
	}
}
