package controller

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	logger := slog.New(h)

	ip := pip.New(nil, logger)
	require.NotNil(t, ip)

	ap := pap.New(nil, logger)
	require.NotNil(t, ap)

	ep := pep.New(nil, logger)
	require.NotNil(t, ep)

	testCases := []struct {
		name        string
		options     []Option
		wantName    string
		wantVersion string
		wantFull    string
		wantCtx     context.Context
		wantLogger  *slog.Logger
		wantPEP     *pep.PEP
		wantPAP     *pap.PAP
		wantPIP     *pip.PIP
		wantMapping []mapping.Mapper
	}{
		{
			name:    "no options",
			wantCtx: context.Background(),
		},
		{
			name:        "name + version",
			options:     []Option{WithNameVersion("x1", "v1")},
			wantName:    "x1",
			wantVersion: "v1",
			wantFull:    "x1 v1",
			wantCtx:     context.Background(),
		},
		{
			name:    "context",
			options: []Option{WithContext(ctx)},
			wantCtx: ctx,
		},
		{
			name:       "logger",
			options:    []Option{WithLogger(logger)},
			wantCtx:    context.Background(),
			wantLogger: logger,
		},
		{
			name:    "pep",
			options: []Option{WithPEP(ep)},
			wantCtx: context.Background(),
			wantPEP: ep,
		},
		{
			name:    "pip",
			options: []Option{WithPIP(ip)},
			wantCtx: context.Background(),
			wantPIP: ip,
		},
		{
			name:    "pap",
			options: []Option{WithPAP(ap)},
			wantCtx: context.Background(),
			wantPAP: ap,
		},
		{
			name:        "mappings",
			options:     []Option{WithMappings(mapping.DoelbindingToPrincipal, mapping.RvvaToPrincipal)},
			wantCtx:     context.Background(),
			wantMapping: []mapping.Mapper{mapping.DoelbindingToPrincipal, mapping.RvvaToPrincipal},
		},
		{
			name:        "all",
			options:     []Option{WithPAP(ap), WithLogger(logger), WithPIP(ip), WithNameVersion("x1", "v1"), WithPEP(ep), WithMappings(mapping.DoelbindingToPrincipal, mapping.RvvaToPrincipal)},
			wantName:    "x1",
			wantVersion: "v1",
			wantFull:    "x1 v1",
			wantCtx:     context.Background(),
			wantLogger:  logger,
			wantPEP:     ep,
			wantPAP:     ap,
			wantPIP:     ip,
			wantMapping: []mapping.Mapper{mapping.DoelbindingToPrincipal, mapping.RvvaToPrincipal},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewBase(tc.options...)
			require.NotNil(t, got)

			assert.Equal(t, tc.wantName, got.Name)
			assert.Equal(t, tc.wantVersion, got.Version)
			assert.Equal(t, tc.wantFull, got.String())
			assert.Equal(t, tc.wantCtx, got.GetContext())
			assert.Equal(t, tc.wantLogger, got.GetLogger())
			assert.Equal(t, tc.wantPEP, got.GetPEP())
			assert.Equal(t, tc.wantPIP, got.GetPIP())
			assert.Equal(t, tc.wantPAP, got.GetPAP())
			assert.Equal(t, len(tc.wantMapping), len(got.mappers))
		})
	}
}
