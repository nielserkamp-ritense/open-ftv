package authentication

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNewBase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	p, err := pip2.New(ctx, log, pip2.WithKeyValueDB(memory.New(), ""), pip2.WithFileStore("../../testdata/pip/users", false))
	require.NoError(t, err)

	testCases := []struct {
		name    string
		opts    []Option
		wantCtx context.Context
		wantLog *slog.Logger
	}{
		{name: "no options"},
		{name: "context", opts: []Option{WithContext(ctx)}, wantCtx: ctx},
		{name: "logger", opts: []Option{WithLogger(log)}, wantLog: log},
		{name: "entities", opts: []Option{WithEntityGetter(p.GetEntity)}},
		{name: "all", opts: []Option{WithLogger(log), WithEntityGetter(p.GetEntity), WithContext(ctx)}, wantCtx: ctx, wantLog: log},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := newBase(tc.opts)
			require.NotNil(t, got)

			assert.NotNil(t, got.ctx)
			assert.NotNil(t, got.log)

			if tc.wantCtx != nil {
				assert.Equal(t, tc.wantCtx, got.ctx)
			}
			if tc.wantLog != nil {
				assert.Equal(t, tc.wantLog, got.log)
			}
		})
	}
}
