package bundles

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestRunner_GatherLists(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		policies       PolicyLister
		attributes     AttributeLister
		entities       EntityLister
		relations      RelationLister
		wantPolicies   int
		wantAttributes int
		wantEntities   int
		wantRelations  int
	}{
		{name: "none"},
		{name: "0 policies", policies: &myPolicyLister{}},
		{name: "3 policies", policies: &myPolicyLister{count: 3}, wantPolicies: 3},
		{name: "0 attributes", attributes: &myAttributeLister{}},
		{name: "2 attributes", attributes: &myAttributeLister{count: 2}, wantAttributes: 2},
		{name: "0 entities", entities: &myEntityLister{}},
		{name: "5 entities", entities: &myEntityLister{count: 5}, wantEntities: 5},
		{name: "0 relations", relations: &myRelationLister{}},
		{name: "7 relations", relations: &myRelationLister{count: 7}, wantRelations: 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			m := NewManager(
				ctx,
				logger,
				WithStageDelay(time.Millisecond),
				WithPolicyLister(tc.policies),
				WithAttributeLister(tc.attributes),
				WithEntityLister(tc.entities),
				WithRelationLister(tc.relations),
			)
			require.NotNil(t, m)

			m.bundles = map[string]*Config{
				"cedar:4": {
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
			}

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
			r.gatherLists("gather")

			require.Equal(t, 1, h.Count())
			require.Equal(t, uint64(1), r.bundleCount)
			require.Equal(t, uint64(2), r.targetCount)
			require.Equal(t, 1, len(r.bundles))

			assert.Equal(t, tc.wantPolicies, len(r.policies))
			assert.Equal(t, tc.wantAttributes, len(r.attributes))
			assert.Equal(t, tc.wantEntities, len(r.entities))
			assert.Equal(t, tc.wantRelations, len(r.relations))
		})
	}
}

type myPolicyLister struct {
	count int
}

func (l *myPolicyLister) Iterate(f models.PolicyIterator) {
	for range l.count {
		f(&models.Policy{})
	}
}

type myAttributeLister struct {
	count int
}

func (l *myAttributeLister) IterateAttributes(f models.AttributeIterator) {
	for range l.count {
		f(&models.Attribute{})
	}
}

type myEntityLister struct {
	count int
}

func (l *myEntityLister) IterateEntities(f models.EntityIterator) {
	for range l.count {
		f(&models.Entity{})
	}
}

type myRelationLister struct {
	count int
}

func (l *myRelationLister) IterateRelations(f models.RelationIterator) {
	for range l.count {
		f(&models.Relation{})
	}
}
