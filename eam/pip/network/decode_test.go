package network

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestDecodeData(t *testing.T) {
	m1 := map[string]interface{}{"hello": "world"}
	m2 := map[string]any{"hello": "world", "int": 123, "bool": true, "type": "xsd:short"}
	m3 := map[string]any{"hello": "mars", "int": 321, "bool": false, "type": "xsd:string"}
	m4 := map[string]any{"hello": "jupiter", "int": true, "bool": 123}
	m5 := map[string]any{"hello": "venus", "int": "2", "bool": true, "type": "xsd:string"}
	m6 := map[string]any{"hello": "uranus", "int": 999, "bool": false, "type": "xsd:anyType"}
	m11 := map[string]any{"t1": "user", "id1": "alice", "t2": "rel", "id2": "knows", "t3": "user", "id3": "bob"}
	m12 := map[string]any{"t1": "user", "id1": "janice", "t2": "rel", "id2": "knows", "t3": "admin", "id3": "danny"}
	m13 := map[string]any{"t1": "admin", "id1": "danny", "t2": "rel", "id2": "knows", "t3": "user", "id3": "alice"}

	s1 := []any{m4, m2, m3}
	s2 := []any{m5, m6}
	s3 := []any{m4, m2}
	s4 := []any{m12, m13, m11}

	mm1 := map[string]any{"first": s1, "second": m1, "third": map[string]any{"sub": s2}}
	mm2 := map[string]any{"second": m1, "third": map[string]any{"sub": m3}, "first": s3}
	mm3 := map[string]any{"first": m2, "second": s4}
	mm4 := map[string]any{"attr": mm1, "rel": mm3, "ent": mm2}

	testCases := []struct {
		name     string
		data     any
		dec      *ResponseMapping
		wantErr  bool
		wantAttr map[string]models.Attribute
		wantEnt  map[string]models.Entity
		wantRel  map[string]models.Relation
	}{
		{
			name: "no data",
			dec: &ResponseMapping{
				Attributes: []*AttributesMapping{
					{Base: "first", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "int", TypeField: "type"}}},
					{Base: "third.sub", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "bool", TypeField: "type"}}},
				},
			},
		},
		{
			name: "dummy decoder",
			data: "anything",
			dec:  &ResponseMapping{},
		},
		{
			name: "attributes",
			data: mm1,
			dec: &ResponseMapping{
				Attributes: []*AttributesMapping{
					{Base: "first", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "int", TypeField: "type"}}},
					{Base: "third.sub", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "bool", TypeField: "type"}}},
				},
			},
			wantAttr: map[string]models.Attribute{
				"world":   models.NewOriginalAttribute("world", int64(123), 123, "xsd:short"),
				"mars":    models.NewOriginalAttribute("mars", "321", 321, "xsd:string"),
				"jupiter": models.NewOriginalAttribute("jupiter", true, true, ""),
				"venus":   models.NewOriginalAttribute("venus", "true", true, "xsd:string"),
				"uranus":  models.NewOriginalAttribute("uranus", "false", false, "xsd:anyType"),
			},
		},
		{
			name: "entities",
			data: mm2,
			dec: &ResponseMapping{
				Entities: []*EntityMapping{
					{Base: "third.sub", TypeField: "hello", IDField: "bool"},
					{Base: "first", TypeField: "hello", IDField: "int"},
				},
			},
			wantEnt: map[string]models.Entity{
				"world::123":    models.NewEntity("world", "123", models.NewAttributeSet()),
				"mars::false":   models.NewEntity("mars", "false", models.NewAttributeSet()),
				"jupiter::true": models.NewEntity("jupiter", "true", models.NewAttributeSet()),
			},
		},
		{
			name: "relations",
			data: map[string]any{"rel": mm3},
			dec: &ResponseMapping{
				Relations: []*RelationMapping{
					{Base: "rel.second", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
				},
			},
			wantRel: map[string]models.Relation{
				"user::alice|rel::knows|user::bob": models.NewRelation(
					models.NewEntity("user", "alice", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("user", "bob", models.NewAttributeSet()),
				),
				"user::janice|rel::knows|admin::danny": models.NewRelation(
					models.NewEntity("user", "janice", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("admin", "danny", models.NewAttributeSet()),
				),
				"admin::danny|rel::knows|user::alice": models.NewRelation(
					models.NewEntity("admin", "danny", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("user", "alice", models.NewAttributeSet()),
				),
			},
		},
		{
			name: "all",
			data: mm4,
			dec: &ResponseMapping{
				Relations: []*RelationMapping{
					{Base: "rel.second", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
				},
				Attributes: []*AttributesMapping{
					{Base: "attr.third.sub", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "bool", TypeField: "type"}}},
					{Base: "attr.first", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "int", TypeField: "type"}}},
				},
				Entities: []*EntityMapping{
					{Base: "ent.first", TypeField: "hello", IDField: "int"},
					{Base: "ent.third.sub", TypeField: "hello", IDField: "bool"},
				},
			},
			wantAttr: map[string]models.Attribute{
				"world":   models.NewOriginalAttribute("world", int64(123), 123, "xsd:short"),
				"mars":    models.NewOriginalAttribute("mars", "321", 321, "xsd:string"),
				"jupiter": models.NewOriginalAttribute("jupiter", true, true, ""),
				"venus":   models.NewOriginalAttribute("venus", "true", true, "xsd:string"),
				"uranus":  models.NewOriginalAttribute("uranus", "false", false, "xsd:anyType"),
			},
			wantEnt: map[string]models.Entity{
				"world::123":    models.NewEntity("world", "123", models.NewAttributeSet()),
				"mars::false":   models.NewEntity("mars", "false", models.NewAttributeSet()),
				"jupiter::true": models.NewEntity("jupiter", "true", models.NewAttributeSet()),
			},
			wantRel: map[string]models.Relation{
				"user::alice|rel::knows|user::bob": models.NewRelation(
					models.NewEntity("user", "alice", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("user", "bob", models.NewAttributeSet()),
				),
				"user::janice|rel::knows|admin::danny": models.NewRelation(
					models.NewEntity("user", "janice", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("admin", "danny", models.NewAttributeSet()),
				),
				"admin::danny|rel::knows|user::alice": models.NewRelation(
					models.NewEntity("admin", "danny", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("user", "alice", models.NewAttributeSet()),
				),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			attr := models.NewAttributeSet()
			ent := models.NewEntitySet()
			rel := models.NewRelationSet(ent)

			m := &manager{logger: logger, attributes: attr, entities: ent, relations: rel, newAttributes: models.NewAttributeSet}

			r := &runner{logger: logger, data: tc.data, manager: m}
			r.decodeData(tc.dec, 200)

			for k := range tc.wantAttr {
				want, got := tc.wantAttr[k], attr.GetAttribute(k)
				require.NotNil(t, got)
				require.EqualValues(t, want, got)
			}

			for k := range tc.wantEnt {
				want, got := tc.wantEnt[k], ent.GetEntity(k)
				require.NotNil(t, got)

				assert.Equal(t, want.UID(), got.UID())
				assert.Equal(t, want.Type(), got.Type())
				assert.Equal(t, want.ID(), got.ID())
				assert.EqualValues(t, want.Parents(), got.Parents())

				want.Attributes().IterateAttributes(func(attr1 models.Attribute) {
					attr2 := got.Attributes().GetAttribute(attr1.Key())
					assert.EqualValues(t, attr1, attr2)
				})

				got.Attributes().IterateAttributes(func(attr1 models.Attribute) {
					attr2 := want.Attributes().GetAttribute(attr1.Key())
					assert.EqualValues(t, attr1, attr2)
				})
			}

			for k := range tc.wantRel {
				want, got := tc.wantRel[k], rel.GetRelation(k)
				require.NotNil(t, got)
				assert.Equal(t, want.UID(), got.UID())
				assert.EqualValues(t, want.Subject(), got.Subject())
				assert.EqualValues(t, want.Predicate(), got.Predicate())
				assert.EqualValues(t, want.Object(), got.Object())
			}
		})
	}
}

func TestDecodeResponse(t *testing.T) {
	const goodYAML = `---
attributes:
  - key: code
    value: 119
    type: xsd:short
  - key: name
    value: "first code"
`

	const goodJSON = `{
"attributes": [
  {  
    "key": "code",
    "value": 119,
    "type": "xsd:short"
  },
  {
    "key": "name",
    "value": "first code"
  }
]}`

	const goodTOML = `
[[attributes]]
key = "code"
value = 119
type = "xsd:short"

[[attributes]]
key = "name"
value = "first code"
`

	testCases := []struct {
		name     string
		data     string
		content  string
		dec      *ResponseMapping
		wantErr  bool
		wantAttr map[string]models.Attribute
		wantEnt  map[string]models.Entity
		wantRel  map[string]models.Relation
	}{
		{
			name:    "no decoder",
			data:    "anything",
			wantErr: true,
		},
		{
			name:    "bad content-type",
			data:    "haha",
			content: "x-bad-data",
			dec:     &ResponseMapping{},
			wantErr: true,
		},
		{
			name:    "no content-type",
			data:    "haha",
			dec:     &ResponseMapping{},
			wantErr: true,
		},
		{
			name:    "bad yaml",
			data:    "- 6 : - : \000",
			content: "application/yaml",
			dec:     &ResponseMapping{},
			wantErr: true,
		},
		{
			name:    "bad toml",
			data:    "- 6 : - : \000",
			content: "application/toml",
			dec:     &ResponseMapping{},
			wantErr: true,
		},
		{
			name:    "bad json",
			data:    "} not json !",
			content: "application/json",
			dec:     &ResponseMapping{},
			wantErr: true,
		},
		{
			name:    "no content-type - bad json",
			data:    "} not json !",
			dec:     &ResponseMapping{},
			wantErr: true,
		},
		{
			name:    "good yaml",
			data:    goodYAML,
			content: "application/yaml",
			dec: &ResponseMapping{
				Attributes: []*AttributesMapping{
					{Base: "attributes", Map: []*AttributeMapping{{KeyField: "key", ValueField: "value", TypeField: "type"}}},
				},
			},
			wantAttr: map[string]models.Attribute{
				"code": models.NewOriginalAttribute("code", int64(119), uint64(119), "xsd:short"),
				"name": models.NewOriginalAttribute("name", "first code", "first code", ""),
			},
		},
		{
			name: "good json",
			data: goodJSON,
			dec: &ResponseMapping{
				Attributes: []*AttributesMapping{
					{Base: "attributes", Map: []*AttributeMapping{{KeyField: "key", ValueField: "value", TypeField: "type"}}},
				},
			},
			wantAttr: map[string]models.Attribute{
				"code": models.NewOriginalAttribute("code", int64(119), 119.0, "xsd:short"),
				"name": models.NewOriginalAttribute("name", "first code", "first code", ""),
			},
		},
		{
			name:    "good toml",
			data:    goodTOML,
			content: "application/toml",
			dec: &ResponseMapping{
				Attributes: []*AttributesMapping{
					{Base: "attributes", Map: []*AttributeMapping{{KeyField: "key", ValueField: "value", TypeField: "type"}}},
				},
			},
			wantAttr: map[string]models.Attribute{
				"code": models.NewOriginalAttribute("code", int64(119), int64(119), "xsd:short"),
				"name": models.NewOriginalAttribute("name", "first code", "first code", ""),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			svc := newService(t, logger, "", "", "", "data", func(req *fiber.Ctx) error {
				if tc.content != "" {
					req.Set("Content-Type", tc.content)
				}
				return req.SendString(tc.data)
			})

			wg := sync.WaitGroup{}
			wg.Add(2)

			go func(wg *sync.WaitGroup) {
				svc.Serve()
				wg.Done()
			}(&wg)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			attr := models.NewAttributeSet()
			ent := models.NewEntitySet()
			rel := models.NewRelationSet(ent)

			m := &manager{ctx: ctx, cancel: cancel, logger: logger, attributes: attr, entities: ent, relations: rel, newAttributes: models.NewAttributeSet}

			req := &Request{
				Name:    "test",
				Method:  "GET",
				URI:     "http://localhost:9900/v1/data",
				Timeout: 100 * time.Millisecond,
				Mapping: tc.dec,
			}

			r := &runner{logger: logger, data: tc.data, manager: m, req: req}

			var execErr error
			go func(wg *sync.WaitGroup) {
				time.Sleep(25 * time.Millisecond)

				execErr = m.execute(logger, r.req)
				svc.Shutdown()
				wg.Done()
			}(&wg)

			wg.Wait()

			if tc.wantErr {
				require.Error(t, execErr)
			} else {
				require.NoError(t, execErr)

				for k := range tc.wantAttr {
					want, got := tc.wantAttr[k], attr.GetAttribute(k)
					require.NotNil(t, got)
					require.True(t, models.AttributeEqual(want, got))
				}

				for k := range tc.wantEnt {
					want, got := tc.wantEnt[k], ent.GetEntity(k)
					require.NotNil(t, got)
					assert.True(t, models.EntityEqual(want, got))
				}

				for k := range tc.wantRel {
					want, got := tc.wantRel[k], rel.GetRelation(k)
					require.NotNil(t, got)
					assert.Equal(t, want.UID(), got.UID())
					assert.True(t, models.EntityEqual(want.Subject(), got.Subject()))
					assert.True(t, models.EntityEqual(want.Predicate(), got.Predicate()))
					assert.True(t, models.EntityEqual(want.Object(), got.Object()))
				}
			}
		})
	}
}
