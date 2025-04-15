package authentication

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

func TestNewBase(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	p := pip.New(ctx, log, pip.WithFileStore("../../../testdata/pip/users", false))
	entities := models.NewEntitySet(p)

	testCases := []struct {
		name         string
		opts         []Option
		wantCtx      context.Context
		wantLog      *slog.Logger
		wantEntities models.EntitySet
	}{
		{name: "no options"},
		{name: "context", opts: []Option{WithContext(ctx)}, wantCtx: ctx},
		{name: "logger", opts: []Option{WithLogger(log)}, wantLog: log},
		{name: "entities", opts: []Option{WithEntities(entities)}, wantEntities: entities},
		{name: "all", opts: []Option{WithLogger(log), WithEntities(entities), WithContext(ctx)}, wantCtx: ctx, wantLog: log, wantEntities: entities},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := newBase(tc.opts)
			require.NotNil(t, got)

			assert.NotNil(t, got.ctx)
			assert.NotNil(t, got.log)
			assert.NotNil(t, got.entities)

			if tc.wantCtx != nil {
				assert.Equal(t, tc.wantCtx, got.ctx)
			}
			if tc.wantLog != nil {
				assert.Equal(t, tc.wantLog, got.log)
			}
			if tc.wantEntities != nil {
				assert.Equal(t, tc.wantEntities, got.entities)
			}
		})
	}
}
