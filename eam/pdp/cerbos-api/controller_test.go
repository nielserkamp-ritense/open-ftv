package cerbos_api

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pdp/controller"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewController(t *testing.T) {
	t.Parallel()

	addr := getAddress()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dummy := slog2.NewDummyHandler(slog.LevelInfo)
	p := pap.New(ctx, slog.New(dummy), pap.WithLanguage("cerbos"), pap.WithFileStore("../../../testdata/policies/cerbos", true))

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
		{
			name:      "good engine - without PAP",
			addr1:     addr,
			addr2:     addr,
			user:      "cerbos",
			pswd:      "cerbos",
			wantSvc1:  true,
			wantSvc2:  true,
			wantInfo:  true,
			wantCount: 1,
		},
		{
			name:      "good engine - with PAP",
			addr1:     addr,
			addr2:     addr,
			user:      "cerbos",
			pswd:      "cerbos",
			opts:      []pdp.Option{pdp.WithPAP(p)},
			wantSvc1:  true,
			wantSvc2:  true,
			wantInfo:  true,
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
