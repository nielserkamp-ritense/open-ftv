//go:build integration

package cerbos_api

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

// TestNewController_GoodEngine covers the good_engine cases from TestNewController that need
// a live Cerbos server (they assert wantInfo, which needs a real server-info RPC to succeed),
// so it only runs with `go test -tags=integration`.
func TestNewController_GoodEngine(t *testing.T) {
	t.Parallel()

	addr := getAddress()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dummy := slog2.NewDummyHandler(slog.LevelInfo)
	p, err := pap.New(ctx, slog.New(dummy), pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cerbos"), pap.WithFileStore("../../../testdata/policies/cerbos", true))
	require.NoError(t, err)

	testCases := []struct {
		name      string
		opts      []pdp.Option
		wantCount int
	}{
		{
			name:      "good engine - without PAP",
			wantCount: 1,
		},
		{
			name:      "good engine - with PAP",
			opts:      []pdp.Option{pdp.WithPAP(p)},
			wantCount: 9,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			opts := append(tc.opts, pdp.WithLogger(logger))

			c := NewController(Config{
				Addr1: addr,
				Addr2: addr,
				User:  "cerbos",
				Pswd:  "cerbos",
			}, opts...)
			require.NotNil(t, c)

			assert.GreaterOrEqual(t, h.Count(), tc.wantCount)

			c2, ok := c.(*controller)
			require.True(t, ok)
			require.NotNil(t, c2)

			require.NotNil(t, c2.engine)
			require.NotNil(t, c2.admin)
			require.NotNil(t, c2.info)
		})
	}
}
