package authentication

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	h := slog2.NewDummyHandler(slog.LevelInfo)
	log := slog.New(h)

	p := pip2.New(ctx, log, pip2.WithFileStore("../../testdata/pip", false))

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
			t.Parallel()

			b := &base{}
			for i := range tc.opts {
				tc.opts[i](b)
			}

			assert.Equal(t, tc.wantCtx, b.ctx)
			assert.Equal(t, tc.wantLog, b.log)
		})
	}
}
