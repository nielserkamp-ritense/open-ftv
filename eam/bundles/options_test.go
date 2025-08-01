package bundles

import (
	"context"
	"log/slog"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		opts           []Option
		wantPath       string
		wantRecurse    bool
		wantPolicies   bool
		wantAttributes bool
		wantEntities   bool
		wantRelations  bool
		wantWorkers    int
		wantDelay      time.Duration
		wantTimeout    time.Duration
	}{
		{
			name:        "config",
			opts:        []Option{WithConfig("/haha", true)},
			wantPath:    "/haha",
			wantRecurse: true,
			wantWorkers: runtime.NumCPU(),
			wantTimeout: time.Minute,
		},
		{
			name:         "policies",
			opts:         []Option{WithPolicyLister(&dummyLister{})},
			wantPolicies: true,
			wantWorkers:  runtime.NumCPU(),
			wantTimeout:  time.Minute,
		},
		{
			name:           "attributes",
			opts:           []Option{WithAttributeLister(&dummyLister{})},
			wantAttributes: true,
			wantWorkers:    runtime.NumCPU(),
			wantTimeout:    time.Minute,
		},
		{
			name:         "entities",
			opts:         []Option{WithEntityLister(&dummyLister{})},
			wantEntities: true,
			wantWorkers:  runtime.NumCPU(),
			wantTimeout:  time.Minute,
		},
		{
			name:          "relations",
			opts:          []Option{WithRelationLister(&dummyLister{})},
			wantRelations: true,
			wantWorkers:   runtime.NumCPU(),
			wantTimeout:   time.Minute,
		},
		{
			name:        "workers",
			opts:        []Option{MaxWorkers(7)},
			wantWorkers: 7,
			wantTimeout: time.Minute,
		},
		{
			name:        "stage delay",
			opts:        []Option{WithStageDelay(3 * time.Second)},
			wantWorkers: runtime.NumCPU(),
			wantDelay:   3 * time.Second,
			wantTimeout: time.Minute,
		},
		{
			name:        "bundle timeout",
			opts:        []Option{BundleTimeout(7 * time.Minute)},
			wantWorkers: runtime.NumCPU(),
			wantTimeout: 7 * time.Minute,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			got := NewManager(ctx, logger, tc.opts...)
			require.NotNil(t, got)

			assert.Equal(t, tc.wantPath, got.path)
			assert.Equal(t, tc.wantRecurse, got.recurse)
			assert.Equal(t, tc.wantPolicies, got.policies != nil)
			assert.Equal(t, tc.wantAttributes, got.attributes != nil)
			assert.Equal(t, tc.wantEntities, got.entities != nil)
			assert.Equal(t, tc.wantRelations, got.relations != nil)
			assert.Equal(t, tc.wantWorkers, got.workers)
			assert.Equal(t, tc.wantDelay, got.stageDelay)
			assert.Equal(t, tc.wantTimeout, got.bundleTimeout)
		})
	}
}

type dummyLister struct{}

func (l *dummyLister) Iterate(models.PolicyIterator)                     {}
func (l *dummyLister) IterateAttributes(models.AttributeIterator)        {}
func (l *dummyLister) IterateEntities(iterator models.EntityIterator)    {}
func (l *dummyLister) IterateRelations(iterator models.RelationIterator) {}
