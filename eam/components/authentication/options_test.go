package authentication

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	log := slog.New(h)

	p := pip.New(ctx, log, pip.WithFileStore("../../../testdata/pip", false))
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
			t.Parallel()

			b := &base{}
			for i := range tc.opts {
				tc.opts[i](b)
			}

			assert.Equal(t, tc.wantCtx, b.ctx)
			assert.Equal(t, tc.wantLog, b.log)
			assert.Equal(t, tc.wantEntities, b.entities)
		})
	}
}
