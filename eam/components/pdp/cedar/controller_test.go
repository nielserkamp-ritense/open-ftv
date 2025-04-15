package cedar

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewController(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		store1   string
		recurse1 bool
		store2   string
		recurse2 bool
		wantLog  int
	}{
		{
			name:    "no stores",
			wantLog: 1,
		},
		{
			name:    "pip store - no recurse",
			store1:  "../../../../testdata/pip",
			wantLog: 1,
		},
		{
			name:     "pip store - recurse",
			store1:   "../../../../testdata/pip",
			recurse1: true,
			wantLog:  1,
		},
		{
			name:    "pap store - no recurse",
			store2:  "../../../../testdata/policies/cedar",
			wantLog: 1,
		},
		{
			name:     "pap store - recurse",
			store2:   "../../../../testdata/policies/cedar",
			recurse2: true,
			wantLog:  6,
		},
		{
			name:     "pip & pap store - recurse",
			store1:   "../../../../testdata/pip",
			recurse1: true,
			store2:   "../../../../testdata/policies/cedar",
			recurse2: true,
			wantLog:  6,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			ip := pip.New(nil, logger, pip.WithFileStore(tc.store1, tc.recurse1))
			require.NotNil(t, ip)

			ap := pap.New(nil, logger, pap.WithLanguage("cedar"), pap.WithFileStore(tc.store2, tc.recurse2))
			require.NotNil(t, ap)

			h.Clear()

			c := NewController(pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
			require.NotNil(t, c)

			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}
