package cerbos_api

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// TestNewController covers the cases that don't need a live Cerbos server. The good_engine
// cases (which do) live in TestNewController_GoodEngine, gated behind the integration tag.
func TestNewController(t *testing.T) {
	t.Parallel()

	addr := getAddress()

	testCases := []struct {
		name      string
		addr1     string
		addr2     string
		ca        string
		user      string
		pswd      string
		opts      []pdp.Option
		wantSvc1  bool
		wantSvc2  bool
		wantInfo  bool
		wantCount int
	}{
		{
			name:      "bad addresses",
			addr1:     "\000\001",
			addr2:     "\002\003",
			user:      "cerbos",
			pswd:      "cerbos",
			wantCount: 1,
		},
		{
			name:      "bad address1",
			addr1:     "\000\001",
			addr2:     addr,
			user:      "cerbos",
			pswd:      "cerbos",
			wantSvc2:  true,
			wantCount: 1,
		},
		{
			name:      "bad address2",
			addr1:     addr,
			addr2:     "\001\002",
			user:      "cerbos",
			pswd:      "cerbos",
			wantSvc1:  true,
			wantCount: 2,
		},
		{
			name:      "invalid engine",
			addr1:     "dns:localhost:9999",
			addr2:     addr,
			user:      "cerbos",
			pswd:      "cerbos",
			wantSvc1:  true,
			wantSvc2:  true,
			wantCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			opts := append(tc.opts, pdp.WithLogger(logger))

			c := NewController(Config{
				Addr1: tc.addr1,
				Addr2: tc.addr2,
				CA:    tc.ca,
				User:  tc.user,
				Pswd:  tc.pswd,
			}, opts...)
			require.NotNil(t, c)

			assert.GreaterOrEqual(t, h.Count(), tc.wantCount)

			c2, ok := c.(*controller)
			require.True(t, ok)
			require.NotNil(t, c2)

			if tc.wantSvc1 {
				require.NotNil(t, c2.engine)
			} else {
				require.Nil(t, c2.engine)
			}

			if tc.wantSvc2 {
				require.NotNil(t, c2.admin)
			} else {
				require.Nil(t, c2.admin)
			}

			if tc.wantInfo {
				require.NotNil(t, c2.info)
			} else {
				require.Nil(t, c2.info)
			}
		})
	}
}
