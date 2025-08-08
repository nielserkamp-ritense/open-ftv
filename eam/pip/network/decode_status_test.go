package network

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestDecodeStatus(t *testing.T) {
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

	testCases := []struct {
		name     string
		data     any
		dec      *ResponseMapping
		status   int
		wantAttr map[string]*models.Attribute
		wantEnt  map[string]*models.Entity
		wantRel  map[string]*models.Relation
	}{
		{
			name: "200 match - no data",
			dec: &ResponseMapping{
				StatusCodes: []*StatusCode{
					{
						Code:        "200",
						Description: "OK",
						Attributes:  &AttributesMapping{Base: "first", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "int", TypeField: "type"}}},
					},
				},
			},
			status: 200,
		},
		{
			name: "200 match - with attributes",
			data: mm1,
			dec: &ResponseMapping{
				StatusCodes: []*StatusCode{
					{
						Code:        "200",
						Description: "OK",
						Attributes:  &AttributesMapping{Base: "first", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "int", TypeField: "type"}}},
					},
				},
			},
			status: 200,
			wantAttr: map[string]*models.Attribute{
				"world":   models.NewOriginalAttribute("world", int64(123), 123, "xsd:short"),
				"mars":    models.NewOriginalAttribute("mars", "321", 321, "xsd:string"),
				"jupiter": models.NewOriginalAttribute("jupiter", true, true, ""),
			},
		},
		{
			name: "201 non-match - with attributes",
			data: mm1,
			dec: &ResponseMapping{
				StatusCodes: []*StatusCode{
					{
						Code:        "200",
						Description: "OK",
						Attributes:  &AttributesMapping{Base: "first", Map: []*AttributeMapping{{KeyField: "hello", ValueField: "int", TypeField: "type"}}},
					},
				},
			},
			status: 201,
		},
		{
			name: "200 match - with entities",
			data: mm2,
			dec: &ResponseMapping{
				StatusCodes: []*StatusCode{
					{
						Code:        "200",
						Description: "OK",
						Entity:      &EntityMapping{Base: "first", TypeField: "hello", IDField: "int"},
					},
				},
			},
			status: 200,
			wantEnt: map[string]*models.Entity{
				"world::123":    models.NewEntity("world", "123", models.NewAttributeSet()),
				"jupiter::true": models.NewEntity("jupiter", "true", models.NewAttributeSet()),
			},
		},
		{
			name: "200 match - with relations",
			data: map[string]any{"rel": mm3},
			dec: &ResponseMapping{
				StatusCodes: []*StatusCode{
					{
						Code:        "200",
						Description: "OK",
						Relation:    &RelationMapping{Base: "rel.second", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
					},
				},
			},
			status: 200,
			wantRel: map[string]*models.Relation{
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

			m := &manager{logger: logger, addAttribute: attr.AddAttribute, addEntity: ent.AddEntity, addRelation: rel.AddRelation}

			r := &runner{logger: logger, data: tc.data, manager: m}
			r.decodeData(tc.dec, tc.status)

			for k := range tc.wantAttr {
				want, got := tc.wantAttr[k], attr.GetAttribute(k)
				require.NotNilf(t, got, k)
				require.EqualValuesf(t, want, got, k)
			}

			for k := range tc.wantEnt {
				want, got := tc.wantEnt[k], ent.GetEntity(k)
				require.NotNilf(t, got, k)

				assert.Equalf(t, want.UID(), got.UID(), k)
				assert.Equalf(t, want.Type(), got.Type(), k)
				assert.Equalf(t, want.ID(), got.ID(), k)
				assert.EqualValuesf(t, want.Parents(), got.Parents(), k)

				want.Attributes().IterateAttributes(func(attr1 *models.Attribute) {
					attr2 := got.Attributes().GetAttribute(attr1.Key())
					assert.EqualValuesf(t, attr1, attr2, k)
				})

				got.Attributes().IterateAttributes(func(attr1 *models.Attribute) {
					attr2 := want.Attributes().GetAttribute(attr1.Key())
					assert.EqualValuesf(t, attr1, attr2, k)
				})
			}

			for k := range tc.wantRel {
				want, got := tc.wantRel[k], rel.GetRelation(k)
				require.NotNilf(t, got, k)
				assert.Equalf(t, want.UID(), got.UID(), k)
				assert.EqualValuesf(t, want.Subject(), got.Subject(), k)
				assert.EqualValuesf(t, want.Predicate(), got.Predicate(), k)
				assert.EqualValuesf(t, want.Object(), got.Object(), k)
			}
		})
	}
}
