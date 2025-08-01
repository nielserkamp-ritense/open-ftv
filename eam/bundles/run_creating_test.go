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

func TestRunner_InitCountsAndSlices(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		bundles     map[string]*Config
		wantBundles uint64
		wantTargets uint64
	}{
		{
			name:    "none",
			bundles: map[string]*Config{},
		},
		{
			name: "1 bundle with 1 target",
			bundles: map[string]*Config{
				"1": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "1",
					Language: "cedar",
					Tags:     []string{"t1"},
					Targets:  []*Target{{URI: "http://localhost:8080/v1/bundle", APIKey: "abcdef", Compress: "gzip"}},
				},
			},
			wantBundles: 1,
			wantTargets: 1,
		},
		{
			name: "1 bundle with 3 targets",
			bundles: map[string]*Config{
				"2": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "2",
					Language: "cedar",
					Tags:     []string{"t1", "t2"},
					Targets: []*Target{
						{URI: "http://localhost:8080/v1/bundle", APIKey: "abcdef", Compress: "gzip"},
						{URI: "http://localhost:8081/v1/bundle", APIKey: "abcdef", Compress: "gzip"},
						{URI: "http://localhost:8082/v1/bundle", APIKey: "abcdef", Compress: "gzip"},
					},
				},
			},
			wantBundles: 1,
			wantTargets: 3,
		},
		{
			name: "2 bundles with 5 targets",
			bundles: map[string]*Config{
				"3": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "3",
					Language: "cedar",
					Tags:     []string{"t1", "t2"},
					Targets: []*Target{
						{URI: "http://localhost:8080/v1/bundle", APIKey: "abcdef", Compress: "gzip"},
						{URI: "http://localhost:8081/v1/bundle", APIKey: "abcdef", Compress: "gzip"},
						{URI: "http://localhost:8082/v1/bundle", APIKey: "abcdef", Compress: "gzip"},
					},
				},
				"4": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "4",
					Language: "cedar",
					Tags:     []string{"t3"},
					Targets: []*Target{
						{URI: "http://localhost:8085/v1/bundle", APIKey: "abcdef", Compress: "gzip"},
						{URI: "http://localhost:8086/v1/bundle", APIKey: "abcdef", Compress: "gzip"},
					},
				},
			},
			wantBundles: 2,
			wantTargets: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			m := NewManager(ctx, logger, WithStageDelay(time.Millisecond))
			require.NotNil(t, m)

			m.bundles = tc.bundles

			client := memory.New()
			require.NotNil(t, client)

			handler := NewPersistence(ctx, client, "")
			require.NotNil(t, handler)

			d, err := handler.Generate("haha")
			require.NoError(t, err)
			require.NotNil(t, d)

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
			r.initCountsAndSlices()

			require.Equal(t, tc.wantBundles, r.bundleCount)
			require.Equal(t, tc.wantTargets, r.targetCount)
			require.Equal(t, len(tc.bundles), len(r.bundles))

			for k := range r.bundles {
				b := r.bundles[k]
				assert.Equal(t, d.version, b.Version)

				cfg := tc.bundles[k]
				assert.Equal(t, cfg.Language, b.Language)
				assert.EqualValues(t, cfg.Tags, b.tags)
			}
		})
	}
}
