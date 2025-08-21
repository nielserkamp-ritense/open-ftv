package bundles

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestRunner_Debug(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		msgs   []string
		params [][]any
		want   int
	}{
		{
			name: "none",
		},
		{
			name: "one, no params",
			msgs: []string{"hello world"},
			want: 1,
		},
		{
			name:   "one, with params",
			msgs:   []string{"hello world"},
			params: [][]any{{1, true, "yo"}},
			want:   1,
		},
		{
			name: "few, no params",
			msgs: []string{"hello", "world", "jupiter", "mars"},
			want: 4,
		},
		{
			name:   "few, with params",
			msgs:   []string{"hello", "world", "jupiter", "mars"},
			params: [][]any{{1, true, "yo"}, {true, -3.1415926535}, {}, {"yo"}},
			want:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			r := &runner{logger: logger, d: &Deployment{version: 1, description: "haha"}}

			for i := range tc.msgs {
				var params []any
				if i < len(tc.params) {
					params = tc.params[i]
				}

				r.debug(tc.msgs[i], params...)
			}

			assert.Equal(t, tc.want, h.Count())
		})
	}
}

func TestRunner_Info(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		msgs   []string
		params [][]any
		want   int
	}{
		{
			name: "none",
		},
		{
			name: "one, no params",
			msgs: []string{"hello world"},
			want: 1,
		},
		{
			name:   "one, with params",
			msgs:   []string{"hello world"},
			params: [][]any{{1, true, "yo"}},
			want:   1,
		},
		{
			name: "few, no params",
			msgs: []string{"hello", "world", "jupiter", "mars"},
			want: 4,
		},
		{
			name:   "few, with params",
			msgs:   []string{"hello", "world", "jupiter", "mars"},
			params: [][]any{{1, true, "yo"}, {true, -3.1415926535}, {}, {"yo"}},
			want:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			r := &runner{logger: logger, d: &Deployment{version: 1, description: "haha"}}

			for i := range tc.msgs {
				var params []any
				if i < len(tc.params) {
					params = tc.params[i]
				}

				r.info(tc.msgs[i], params...)
			}

			assert.Equal(t, tc.want, h.Count())
		})
	}
}

func TestRunner_Warn(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		msgs   []string
		params [][]any
		want   int
	}{
		{
			name: "none",
		},
		{
			name: "one, no params",
			msgs: []string{"hello world"},
			want: 1,
		},
		{
			name:   "one, with params",
			msgs:   []string{"hello world"},
			params: [][]any{{1, true, "yo"}},
			want:   1,
		},
		{
			name: "few, no params",
			msgs: []string{"hello", "world", "jupiter", "mars"},
			want: 4,
		},
		{
			name:   "few, with params",
			msgs:   []string{"hello", "world", "jupiter", "mars"},
			params: [][]any{{1, true, "yo"}, {true, -3.1415926535}, {}, {"yo"}},
			want:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			r := &runner{logger: logger, d: &Deployment{version: 1, description: "haha"}}

			for i := range tc.msgs {
				var params []any
				if i < len(tc.params) {
					params = tc.params[i]
				}

				r.warn(tc.msgs[i], params...)
			}

			assert.Equal(t, tc.want, h.Count())
		})
	}
}

func TestRunner_Error(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		msgs   []string
		params [][]any
		want   int
	}{
		{
			name: "none",
		},
		{
			name: "one, no params",
			msgs: []string{"hello world"},
			want: 1,
		},
		{
			name:   "one, with params",
			msgs:   []string{"hello world"},
			params: [][]any{{1, true, "yo"}},
			want:   1,
		},
		{
			name: "few, no params",
			msgs: []string{"hello", "world", "jupiter", "mars"},
			want: 4,
		},
		{
			name:   "few, with params",
			msgs:   []string{"hello", "world", "jupiter", "mars"},
			params: [][]any{{1, true, "yo"}, {true, -3.1415926535}, {}, {"yo"}},
			want:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			r := &runner{logger: logger, d: &Deployment{version: 1, description: "haha"}}

			for i := range tc.msgs {
				var params []any
				if i < len(tc.params) {
					params = tc.params[i]
				}

				r.error(tc.msgs[i], params...)
			}

			assert.Equal(t, tc.want, h.Count())
		})
	}
}

func TestRunner_DummyRun(t *testing.T) {
	t.Parallel()

	t.Run("dummy run", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		m := NewManager(ctx, logger, WithStageDelay(time.Millisecond))
		require.NotNil(t, m)

		client := memory.New()
		require.NotNil(t, client)

		handler := NewKeyValueDB(client, "")
		require.NotNil(t, handler)

		d, err := handler.Generate(ctx, "haha", "")
		require.NoError(t, err)
		require.NotNil(t, d)

		h.Clear()

		r := m.Run(d, handler).(*runner)
		require.NotNil(t, r)

		for r.d.status < Failed {
			time.Sleep(10 * time.Millisecond)
		}

		assert.Equal(t, Completed, r.d.status)
		assert.GreaterOrEqual(t, h.Count(), 15)
	})
}

func TestRunner_RunOnFailed(t *testing.T) {
	t.Parallel()

	t.Run("run on failed", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		m := NewManager(ctx, logger, WithStageDelay(time.Millisecond))
		require.NotNil(t, m)

		client := memory.New()
		require.NotNil(t, client)

		handler := NewKeyValueDB(client, "")
		require.NotNil(t, handler)

		d, err := handler.Generate(ctx, "haha", "")
		require.NoError(t, err)
		require.NotNil(t, d)

		d.status = Failed

		h.Clear()

		r := &runner{
			ctx:           ctx,
			cancel:        cancel,
			m:             m,
			handler:       handler,
			d:             d,
			logger:        m.logger,
			bundleTimeout: time.Second,
			client:        m.client,
		}
		r.run()

		assert.Equal(t, Failed, r.d.status)
		assert.Zero(t, h.Count())
	})
}

func TestRunner_RunOnCompleted(t *testing.T) {
	t.Parallel()

	t.Run("run on completed", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		m := NewManager(ctx, logger, WithStageDelay(time.Millisecond))
		require.NotNil(t, m)

		client := memory.New()
		require.NotNil(t, client)

		handler := NewKeyValueDB(client, "")
		require.NotNil(t, handler)

		d, err := handler.Generate(ctx, "haha", "")
		require.NoError(t, err)
		require.NotNil(t, d)

		d.status = Completed

		h.Clear()

		r := &runner{
			ctx:           ctx,
			cancel:        cancel,
			m:             m,
			handler:       handler,
			d:             d,
			logger:        m.logger,
			bundleTimeout: time.Second,
			client:        m.client,
		}
		r.run()

		assert.Equal(t, Completed, r.d.status)
		assert.Zero(t, h.Count())
	})
}

func TestRunner_RunOnBadStatus(t *testing.T) {
	t.Parallel()

	t.Run("run on bad status", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		m := NewManager(ctx, logger, WithStageDelay(time.Millisecond))
		require.NotNil(t, m)

		client := memory.New()
		require.NotNil(t, client)

		handler := NewKeyValueDB(client, "")
		require.NotNil(t, handler)

		d, err := handler.Generate(ctx, "haha", "")
		require.NoError(t, err)
		require.NotNil(t, d)

		d.status = 99

		h.Clear()

		r := &runner{
			ctx:           ctx,
			cancel:        cancel,
			m:             m,
			handler:       handler,
			d:             d,
			logger:        m.logger,
			bundleTimeout: time.Second,
			client:        m.client,
		}
		r.run()

		assert.Equal(t, Status(99), r.d.status)
		assert.Equal(t, 1, h.Count())
	})
}

func TestRunner_Run_BadVersion(t *testing.T) {
	t.Parallel()

	t.Run("run on bad version", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		m := NewManager(ctx, logger, WithStageDelay(time.Millisecond))
		require.NotNil(t, m)

		client := memory.New()
		require.NotNil(t, client)

		handler := NewKeyValueDB(client, "")
		require.NotNil(t, handler)

		h.Clear()

		r := &runner{
			ctx:           ctx,
			cancel:        cancel,
			m:             m,
			handler:       handler,
			d:             &Deployment{version: 1, description: "yo", status: Creating},
			logger:        m.logger,
			bundleTimeout: time.Second,
			client:        m.client,
		}
		r.run()

		assert.Equal(t, Creating, r.d.status)
		assert.Equal(t, 3, h.Count())
	})
}
