package bundles

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestRunner_GatherTargets(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		policies    []*models.Policy
		attributes  []*models.Attribute
		entities    []*models.Entity
		relations   []*models.Relation
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
			policies:    []*models.Policy{myP2, myP4, myP1, myP5, myP3},
			attributes:  []*models.Attribute{myA1, myA2},
			entities:    []*models.Entity{myU2},
			relations:   []*models.Relation{myR1},
			wantBundles: 1,
			wantTargets: 1,
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
			policies:    []*models.Policy{myP2, myP4, myP1, myP5, myP3},
			attributes:  []*models.Attribute{myA3, myA1, myA4, myA2},
			entities:    []*models.Entity{myU2, myU3, myU1},
			relations:   []*models.Relation{myR1, myR2},
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

			opts := []Option{WithStageDelay(time.Minute), MaxWorkers(2)}
			if tc.policies != nil {
				opts = append(opts, WithPolicyLister(&myLists{policies: tc.policies}))
			}
			if tc.attributes != nil {
				opts = append(opts, WithAttributeLister(&myLists{attributes: tc.attributes}))
			}
			if tc.entities != nil {
				opts = append(opts, WithEntityLister(&myLists{entities: tc.entities}))
			}
			if tc.relations != nil {
				opts = append(opts, WithRelationLister(&myLists{relations: tc.relations}))
			}

			m := NewManager(ctx, logger, opts...)
			require.NotNil(t, m)

			m.bundles = tc.bundles

			client := memory.New()
			require.NotNil(t, client)

			handler := NewKeyValueDB(client, "")
			require.NotNil(t, handler)

			d, err := handler.Generate(ctx, "haha", "")
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
				bundleTimeout: m.bundleTimeout,
				client:        m.client,
			}
			r.gatherTargets("bundle")

			assert.Equal(t, tc.wantBundles, r.bundleCount)
			assert.Equal(t, tc.wantTargets, r.targetCount)
			assert.Equal(t, int(tc.wantTargets), len(r.targets))
		})
	}
}

func TestRunner_SendBundles(t *testing.T) {
	pdp := make([]*httptest.Server, 5)
	for i := range pdp {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("api-key")
			if key != "abcdef" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ct := CompressionTypeFromString(r.Header.Get("Content-Encoding"))
			if ct != CompressGZ && ct != CompressBZ2 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)

			b, _ := json.Marshal(&bundles.BundleActivated{PreviousVersion: 0})
			_, _ = w.Write(b)
		}))

		pdp[i] = srv
	}

	testCases := []struct {
		name        string
		policies    []*models.Policy
		attributes  []*models.Attribute
		entities    []*models.Entity
		relations   []*models.Relation
		bundles     map[string]*Config
		workers     int
		wantErr     bool
		wantLog     int
		wantBundles uint64
		wantTargets uint64
		wantBundled uint64
		wantSent    uint64
	}{
		{
			name:    "none",
			bundles: map[string]*Config{},
			workers: 2,
			wantLog: 4,
		},
		{
			name: "1 bundle, 1 target, only version, gzip",
			bundles: map[string]*Config{
				"1": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "1",
					Language: "cedar",
					Tags:     []string{"t1"},
					Targets:  []*Target{{URI: pdp[0].URL, APIKey: "abcdef", Compress: "gzip"}},
				},
			},
			workers:     2,
			wantLog:     4,
			wantBundles: 1,
			wantTargets: 1,
			wantBundled: 1,
			wantSent:    1,
		},
		{
			name: "1 bundle, 1 target, only version, bzip2",
			bundles: map[string]*Config{
				"1": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "1",
					Language: "cedar",
					Tags:     []string{"t1"},
					Targets:  []*Target{{URI: pdp[0].URL, APIKey: "abcdef", Compress: "bz2"}},
				},
			},
			workers:     2,
			wantLog:     4,
			wantBundles: 1,
			wantTargets: 1,
			wantBundled: 1,
			wantSent:    1,
		},
		{
			name: "1 bundle, 1 target, bad compression (still works)",
			bundles: map[string]*Config{
				"1": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "1",
					Language: "cedar",
					Tags:     []string{"t1"},
					Targets:  []*Target{{URI: pdp[0].URL, APIKey: "abcdef", Compress: "oops"}},
				},
			},
			workers:     2,
			wantLog:     4,
			wantBundles: 1,
			wantTargets: 1,
			wantBundled: 1,
			wantSent:    1,
		},
		{
			name: "1 bundle, 1 target, bad api-key",
			bundles: map[string]*Config{
				"1": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "1",
					Language: "cedar",
					Tags:     []string{"t1"},
					Targets:  []*Target{{URI: pdp[0].URL, APIKey: "wbcdef", Compress: "gzip"}},
				},
			},
			workers:     2,
			wantErr:     true,
			wantLog:     4,
			wantBundles: 1,
			wantTargets: 1,
			wantBundled: 1,
		},
		{
			name:       "2 bundles, 5 targets, all content, gzip",
			policies:   []*models.Policy{myP2, myP4, myP1, myP5, myP3},
			attributes: []*models.Attribute{myA3, myA1, myA4, myA2},
			entities:   []*models.Entity{myU2, myU3, myU1},
			relations:  []*models.Relation{myR1, myR2},
			bundles: map[string]*Config{
				"1": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "1",
					Language: "cedar",
					Tags:     []string{"t3"},
					Targets: []*Target{
						{URI: pdp[0].URL, APIKey: "abcdef", Compress: "gzip"},
						{URI: pdp[1].URL, APIKey: "abcdef", Compress: "bz1"},
						{URI: pdp[4].URL, APIKey: "abcdef", Compress: "oops"},
					},
				},
				"2": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "2",
					Language: "cedar",
					Tags:     []string{"t4"},
					Targets: []*Target{
						{URI: pdp[2].URL, APIKey: "abcdef", Compress: "bzip2"},
						{URI: pdp[3].URL, APIKey: "abcdef", Compress: "gz"},
					},
				},
			},
			workers:     2,
			wantLog:     4,
			wantBundles: 2,
			wantTargets: 5,
			wantBundled: 2,
			wantSent:    5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			opts := []Option{BundleTimeout(time.Minute), MaxWorkers(tc.workers)}
			if tc.policies != nil {
				opts = append(opts, WithPolicyLister(&myLists{policies: tc.policies}))
			}
			if tc.attributes != nil {
				opts = append(opts, WithAttributeLister(&myLists{attributes: tc.attributes}))
			}
			if tc.entities != nil {
				opts = append(opts, WithEntityLister(&myLists{entities: tc.entities}))
			}
			if tc.relations != nil {
				opts = append(opts, WithRelationLister(&myLists{relations: tc.relations}))
			}

			m := NewManager(ctx, logger, opts...)
			require.NotNil(t, m)

			m.bundles = tc.bundles

			client := memory.New()
			require.NotNil(t, client)

			handler := NewKeyValueDB(client, "")
			require.NotNil(t, handler)

			d, err := handler.Generate(ctx, "haha", "")
			require.NoError(t, err)
			require.NotNil(t, d)

			for d.Status() < Sending {
				d, err = handler.Advance()
				require.NoError(t, err)
				require.NotNil(t, d)
			}

			h.Clear()

			r := &runner{
				ctx:           ctx,
				cancel:        cancel,
				m:             m,
				handler:       handler,
				d:             d,
				logger:        m.logger,
				bundleTimeout: m.bundleTimeout,
				client:        m.client,
			}
			r.sending()

			d, err = handler.LastDeployment(ctx)
			require.NoError(t, err)
			require.NotNil(t, d)

			assert.Equal(t, tc.wantLog, h.Count())
			assert.Equal(t, tc.wantBundles, r.bundleCount)
			assert.Equal(t, tc.wantTargets, r.targetCount)
			assert.Equal(t, tc.wantBundled, r.bundledCount)
			assert.Equal(t, tc.wantSent, r.sendCount)

			if tc.wantErr {
				require.Equal(t, Failed, d.Status())
			} else {
				require.Equal(t, Completed, d.Status())
			}
		})
	}
}

func TestSendBundle_BadURL(t *testing.T) {
	t.Parallel()

	t.Run("send bundle - bad url", func(t *testing.T) {
		job := &sendJob{
			key:    "x",
			target: &Target{URI: "\001\002"},
			bundle: &Bundle{
				Version:    1,
				Language:   "cedar",
				Policies:   make(map[string]*policies.Policy),
				Attributes: make(map[string]*attributes.Attribute),
				Entities:   make(map[string]*attributes.Entity),
				Relations:  make(map[string]*attributes.Relation),
			},
			timeout: time.Second,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		count, err := sendBundle(ctx, job)
		require.Error(t, err)
		require.Zero(t, count)
	})
}

func TestSendBundle_NoServer(t *testing.T) {
	t.Parallel()

	t.Run("send bundle - no server", func(t *testing.T) {
		job := &sendJob{
			key:    "x",
			target: &Target{URI: "http://localhost:44444/v1/bundle"},
			bundle: &Bundle{
				Version:    1,
				Language:   "cedar",
				Policies:   make(map[string]*policies.Policy),
				Attributes: make(map[string]*attributes.Attribute),
				Entities:   make(map[string]*attributes.Entity),
				Relations:  make(map[string]*attributes.Relation),
			},
			timeout: time.Second,
			client:  &http.Client{Timeout: time.Second},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		count, err := sendBundle(ctx, job)
		require.Error(t, err)
		require.Zero(t, count)
	})
}

func TestSendBundle_InvalidStatus(t *testing.T) {
	t.Parallel()

	t.Run("send bundle - invalid status", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		job := &sendJob{
			key:    "x",
			target: &Target{URI: srv.URL},
			bundle: &Bundle{
				Version:    1,
				Language:   "cedar",
				Policies:   make(map[string]*policies.Policy),
				Attributes: make(map[string]*attributes.Attribute),
				Entities:   make(map[string]*attributes.Entity),
				Relations:  make(map[string]*attributes.Relation),
			},
			timeout: time.Second,
			client:  &http.Client{Timeout: time.Second},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		count, err := sendBundle(ctx, job)
		require.Error(t, err)
		require.Zero(t, count)
	})
}

func TestSendBundle_InvalidResponse(t *testing.T) {
	t.Parallel()

	t.Run("send bundle - invalid response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte{1, 2, 3, 4})
		}))
		defer srv.Close()

		job := &sendJob{
			key:    "x",
			target: &Target{URI: srv.URL},
			bundle: &Bundle{
				Version:    1,
				Language:   "cedar",
				Policies:   make(map[string]*policies.Policy),
				Attributes: make(map[string]*attributes.Attribute),
				Entities:   make(map[string]*attributes.Entity),
				Relations:  make(map[string]*attributes.Relation),
			},
			timeout: time.Second,
			client:  &http.Client{Timeout: time.Second},
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		count, err := sendBundle(ctx, job)
		require.Error(t, err)
		require.Zero(t, count)
	})
}
