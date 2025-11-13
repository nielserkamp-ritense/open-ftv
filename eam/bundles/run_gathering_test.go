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
		policies       PolicyHandler
		data           DataHandler
		wantPolicies   int
		wantAttributes int
		wantEntities   int
		wantRelations  int
	}{
		{name: "none"},
		{name: "0 policies", policies: &myPolicyHandler{}},
		{name: "3 policies", policies: &myPolicyHandler{count: 3}, wantPolicies: 3},
		{name: "0 attributes", data: &myDataHandler{}},
		{name: "2 attributes", data: &myDataHandler{count1: 2}, wantAttributes: 2},
		{name: "0 entities", data: &myDataHandler{}},
		{name: "5 entities", data: &myDataHandler{count2: 5}, wantEntities: 5},
		{name: "0 relations", data: &myDataHandler{}},
		{name: "7 relations", data: &myDataHandler{count3: 7}, wantRelations: 7},
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
				WithPolicyHandler(tc.policies),
				WithDataHandler(tc.data),
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
						{URI: "http://localhost:8085/v1/bundle", APIKey: "abcdef", Encoding: "gzip"},
						{URI: "http://localhost:8086/v1/bundle", APIKey: "abcdef", Encoding: "gzip"},
					},
				},
			}

			client := memory.New()
			require.NotNil(t, client)

			handler := NewKeyValueDB(client, "")
			require.NotNil(t, handler)

			d, err := handler.Generate(ctx, "h1", "haha", "")
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
				done:          make(chan struct{}),
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

type myPolicyHandler struct {
	count int
}

func (l *myPolicyHandler) Iterate(f models.PolicyIterator) {
	for range l.count {
		f(&models.Policy{})
	}
}

func (l *myPolicyHandler) Create(*models.Policy, string) (*models.Policy, error) { return nil, nil }

type myDataHandler struct {
	count1 int
	count2 int
	count3 int
}

func (l *myDataHandler) IterateAttributes(f models.AttributeIterator) {
	for range l.count1 {
		f(&models.Attribute{})
	}
}

func (l *myDataHandler) IterateEntities(f models.EntityIterator) {
	for range l.count2 {
		f(&models.Entity{})
	}
}

func (l *myDataHandler) IterateRelations(f models.RelationIterator) {
	for range l.count3 {
		f(&models.Relation{})
	}
}

func (l *myDataHandler) AddAttribute(*models.Attribute) (*models.Attribute, error) { return nil, nil }
func (l *myDataHandler) AddEntity(*models.Entity) (*models.Entity, error)          { return nil, nil }
func (l *myDataHandler) AddRelation(*models.Relation) (*models.Relation, error)    { return nil, nil }
