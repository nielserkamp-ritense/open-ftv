package openfga_embedded

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
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
			store1:  "../../../testdata/pip",
			wantLog: 1,
		},
		{
			name:     "pip store - recurse",
			store1:   "../../../testdata/pip",
			recurse1: true,
			wantLog:  1,
		},
		{
			name:    "pap store - no recurse",
			store2:  "../../../testdata/policies/openfga",
			wantLog: 3,
		},
		{
			name:     "pap store - recurse",
			store2:   "../../../testdata/policies/openfga",
			recurse2: true,
			wantLog:  3,
		},
		{
			name:     "pip & pap store - recurse",
			store1:   "../../../testdata/pip",
			recurse1: true,
			store2:   "../../../testdata/policies/openfga",
			recurse2: true,
			wantLog:  3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			ip, err := pip2.New(t.Context(), logger, pip2.WithKeyValueDB(memory.New(), ""), pip2.WithFileStore(tc.store1, tc.recurse1))
			require.NoError(t, err)
			require.NotNil(t, ip)

			ap, err := pap2.New(t.Context(), logger, pap2.WithKeyValueDB(memory.New(), ""), pap2.WithLanguage("openfga"), pap2.WithFileStore(tc.store2, tc.recurse2))
			require.NoError(t, err)
			require.NotNil(t, ap)

			h.Clear()

			c := NewController(pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
			require.NotNil(t, c)

			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}
