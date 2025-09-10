package bundles

import (
	"bytes"
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

func TestRunner_CreateBundles(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		policies       []*models.Policy
		attributes     []*models.Attribute
		entities       []*models.Entity
		relations      []*models.Relation
		wantPolicies   int
		wantAttributes int
		wantEntities   int
		wantRelations  int
	}{
		{name: "none"},
		{name: "0 policies", policies: []*models.Policy{}},
		{name: "few policies, none match", policies: []*models.Policy{myP1, myP2, myP3}},
		{name: "few policies, 2 match", policies: []*models.Policy{myP4, myP1, myP5}, wantPolicies: 2},
		{name: "0 attributes", attributes: []*models.Attribute{}},
		{name: "few attributes, none match", attributes: []*models.Attribute{myA1, myA2}},
		{name: "few attributes, 2 match", attributes: []*models.Attribute{myA3, myA1, myA4, myA2}, wantAttributes: 2},
		{name: "0 entities", entities: []*models.Entity{}},
		{name: "one entity, none match", entities: []*models.Entity{myU2}},
		{name: "few entities, 2 match", entities: []*models.Entity{myU2, myU3, myU1}, wantEntities: 2},
		{name: "0 relations", relations: []*models.Relation{}},
		{name: "1 relation, no match", relations: []*models.Relation{myR1}},
		{name: "few relations, 1 match", relations: []*models.Relation{myR1, myR2}, wantRelations: 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			opts := []Option{WithStageDelay(time.Millisecond)}
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

			m.bundles = map[string]*Config{
				"cedar:4": {
					Policies: true,
					Data:     true,
					Version:  true,
					ID:       "4",
					Language: "cedar",
					Tags:     []string{"t4"},
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
				bundleTimeout: time.Second,
				client:        m.client,
				done:          make(chan struct{}),
			}
			r.createBundles("bundles")

			require.Equal(t, uint64(1), r.bundleCount)
			require.Equal(t, uint64(2), r.targetCount)
			require.Equal(t, 1, len(r.bundles))
			require.NotEmpty(t, r.gitHash)

			b := r.bundles["cedar:4"]
			require.NotNil(t, b)

			assert.Equal(t, tc.wantPolicies, len(b.Policies))
			assert.Equal(t, tc.wantAttributes, len(b.Attributes))
			assert.Equal(t, tc.wantEntities, len(b.Entities))
			assert.Equal(t, tc.wantRelations, len(b.Relations))
		})
	}
}

type myLists struct {
	policies   []*models.Policy
	attributes []*models.Attribute
	entities   []*models.Entity
	relations  []*models.Relation
}

func (l *myLists) Iterate(f models.PolicyIterator) {
	for i := range l.policies {
		f(l.policies[i])
	}
}

func (l *myLists) IterateAttributes(f models.AttributeIterator) {
	for i := range l.attributes {
		f(l.attributes[i])
	}
}

func (l *myLists) IterateEntities(f models.EntityIterator) {
	for i := range l.entities {
		f(l.entities[i])
	}
}

func (l *myLists) IterateRelations(f models.RelationIterator) {
	for i := range l.relations {
		f(l.relations[i])
	}
}

var (
	myP1, _ = models.NewPolicyFromData("1", "cedar", "", "", bytes.NewBufferString("allow();"))
	myP2, _ = models.NewPolicyFromData("2", "cedar", "", "", bytes.NewBufferString("allow();"))
	myP3, _ = models.NewPolicyFromData("3", "cedar", "", "", bytes.NewBufferString("allow();"))
	myP4, _ = models.NewPolicyFromData("4", "cedar", "", "", bytes.NewBufferString("allow();"))
	myP5, _ = models.NewPolicyFromData("5", "cedar", "", "", bytes.NewBufferString("allow();"))

	myA1 = models.NewAttribute("myA1", "hello world")
	myA2 = models.NewAttribute("myA2", true)
	myA3 = models.NewAttribute("myA3", -1111)
	myA4 = models.NewAttribute("myA4", 1.3456)

	myU1 = models.NewEntity("user", "bob", nil)
	myU2 = models.NewEntity("user", "alice", nil)
	myU3 = models.NewEntity("user", "xander", nil)

	myR1 = models.NewRelation(myU1, models.NewEntity("action", "read", nil), models.NewEntity("resource", "book1", nil))
	myR2 = models.NewRelation(myU2, models.NewEntity("action", "read", nil), models.NewEntity("resource", "book1", nil))
)

func init() {
	myP1.WithTags("t1", "t2")
	myP2.WithTags("t1", "t3")
	myP3.WithTags("t2", "t3")
	myP4.WithTags("t1", "t4")
	myP5.WithTags("t4", "t3")

	myA1.WithTags("t1", "t2")
	myA2.WithTags("t3", "t2")
	myA3.WithTags("t2", "t4")
	myA4.WithTags("t4", "t1")

	myU1.WithTags("t4", "t2")
	myU2.WithTags("t3", "t2")
	myU3.WithTags("t1", "t4")

	myR1.WithTags("t1", "t2")
	myR2.WithTags("t4", "t2")
}
