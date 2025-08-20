package network

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestProcessRelation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		sType     string
		sID       string
		pType     string
		pID       string
		oType     string
		oID       string
		obj       *RelationMapping
		wantCount int
		wantKey   string
		wantValue *models.Relation
	}{
		{name: "no subject type", sID: "alice", pType: "rel", pID: "knows", oType: "user", oID: "bob", obj: &RelationMapping{Base: "first"}, wantCount: 1},
		{name: "no subject id", sType: "user", pType: "rel", pID: "knows", oType: "user", oID: "bob", obj: &RelationMapping{Base: "first"}, wantCount: 1},
		{name: "no predicate type", sType: "user", sID: "alice", pID: "knows", oType: "user", oID: "bob", obj: &RelationMapping{Base: "first"}, wantCount: 1},
		{name: "no predicate id", sType: "user", sID: "alice", pType: "rel", oType: "user", oID: "bob", obj: &RelationMapping{Base: "first"}, wantCount: 1},
		{name: "no object type", sType: "user", sID: "alice", pType: "rel", pID: "knows", oID: "bob", obj: &RelationMapping{Base: "first"}, wantCount: 1},
		{name: "no object id", sType: "user", sID: "alice", pType: "rel", pID: "knows", oType: "user", obj: &RelationMapping{Base: "first"}, wantCount: 1},
		{
			name:    "all ok",
			sType:   "user",
			sID:     "alice",
			pType:   "rel",
			pID:     "knows",
			oType:   "user",
			oID:     "bob",
			obj:     &RelationMapping{Base: "first"},
			wantKey: "user::alice|rel::knows|user::bob",
			wantValue: models.NewRelation(
				models.NewEntity("user", "alice", models.NewAttributeSet()),
				models.NewEntity("rel", "knows", models.NewAttributeSet()),
				models.NewEntity("user", "bob", models.NewAttributeSet()),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ent := models.NewEntitySet()
			addE := func(in *models.Entity) (*models.Entity, error) {
				ent.AddEntity(in)
				return in, nil
			}

			rel := models.NewRelationSet(ent)
			addR := func(in *models.Relation) (*models.Relation, error) {
				rel.AddRelation(in)
				return in, nil
			}

			r := &runner{logger: logger, manager: &manager{logger: logger, addEntity: addE, addRelation: addR}}
			r.processRelation(tc.sType, tc.sID, tc.pType, tc.pID, tc.oType, tc.oID, tc.obj)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				got := rel.GetRelation(tc.wantKey)
				assert.EqualValues(t, tc.wantValue, got)
			}
		})
	}
}

func TestDecodeRelationMap(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{"t1": "user", "id1": "alice", "t2": "rel", "id2": "knows", "t3": "user", "id3": "bob"}

	testCases := []struct {
		name      string
		m         map[string]any
		obj       *RelationMapping
		wantCount int
		wantKey   string
		wantValue *models.Relation
	}{
		{
			name:      "no subject type code, no subject type value",
			m:         m1,
			obj:       &RelationMapping{Base: "first", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantCount: 1,
		},
		{
			name:      "invalid type code, no type value",
			m:         m1,
			obj:       &RelationMapping{Base: "first", SubjectTypeField: "abc", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantCount: 1,
		},
		{
			name:    "no type code, type value",
			m:       m1,
			obj:     &RelationMapping{Base: "first", SubjectTypeValue: "admin", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantKey: "admin::alice|rel::knows|user::bob",
			wantValue: models.NewRelation(
				models.NewEntity("admin", "alice", models.NewAttributeSet()),
				models.NewEntity("rel", "knows", models.NewAttributeSet()),
				models.NewEntity("user", "bob", models.NewAttributeSet()),
			),
		},
		{
			name:    "invalid type code, type value",
			m:       m1,
			obj:     &RelationMapping{Base: "first", SubjectTypeField: "abc", SubjectTypeValue: "admin", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantKey: "admin::alice|rel::knows|user::bob",
			wantValue: models.NewRelation(
				models.NewEntity("admin", "alice", models.NewAttributeSet()),
				models.NewEntity("rel", "knows", models.NewAttributeSet()),
				models.NewEntity("user", "bob", models.NewAttributeSet()),
			),
		},
		{
			name:    "valid type code, type value",
			m:       m1,
			obj:     &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectTypeValue: "admin", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantKey: "user::alice|rel::knows|user::bob",
			wantValue: models.NewRelation(
				models.NewEntity("user", "alice", models.NewAttributeSet()),
				models.NewEntity("rel", "knows", models.NewAttributeSet()),
				models.NewEntity("user", "bob", models.NewAttributeSet()),
			),
		},
		{
			name:      "empty map",
			m:         map[string]any{},
			obj:       &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantCount: 1,
		},
		{
			name:      "no subject id code, no subject id value",
			m:         m1,
			obj:       &RelationMapping{Base: "first", SubjectTypeField: "t1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantCount: 1,
		},
		{
			name:    "no id code, id value",
			m:       m1,
			obj:     &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDValue: "jasper", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantKey: "user::jasper|rel::knows|user::bob",
			wantValue: models.NewRelation(
				models.NewEntity("user", "jasper", models.NewAttributeSet()),
				models.NewEntity("rel", "knows", models.NewAttributeSet()),
				models.NewEntity("user", "bob", models.NewAttributeSet()),
			),
		},
		{
			name:      "invalid id code, no id value",
			m:         m1,
			obj:       &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "oops", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantCount: 1,
		},
		{
			name:    "invalid id code, id value",
			m:       m1,
			obj:     &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "nope", SubjectIDValue: "henk", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantKey: "user::henk|rel::knows|user::bob",
			wantValue: models.NewRelation(
				models.NewEntity("user", "henk", models.NewAttributeSet()),
				models.NewEntity("rel", "knows", models.NewAttributeSet()),
				models.NewEntity("user", "bob", models.NewAttributeSet()),
			),
		},
		{
			name:    "valid id code, id value",
			m:       m1,
			obj:     &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "id1", SubjectIDValue: "pieter", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			wantKey: "user::alice|rel::knows|user::bob",
			wantValue: models.NewRelation(
				models.NewEntity("user", "alice", models.NewAttributeSet()),
				models.NewEntity("rel", "knows", models.NewAttributeSet()),
				models.NewEntity("user", "bob", models.NewAttributeSet()),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ent := models.NewEntitySet()
			addE := func(in *models.Entity) (*models.Entity, error) {
				ent.AddEntity(in)
				return in, nil
			}

			rel := models.NewRelationSet(ent)
			addR := func(in *models.Relation) (*models.Relation, error) {
				rel.AddRelation(in)
				return in, nil
			}

			r := &runner{logger: logger, manager: &manager{logger: logger, addEntity: addE, addRelation: addR}}
			r.decodeRelationMap(tc.m, tc.obj)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				got := rel.GetRelation(tc.wantKey)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantValue.UID(), got.UID())
				assert.EqualValues(t, tc.wantValue.Subject(), got.Subject())
				assert.EqualValues(t, tc.wantValue.Predicate(), got.Predicate())
				assert.EqualValues(t, tc.wantValue.Object(), got.Object())
			}
		})
	}
}

func TestDecodeRelationData(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{"t1": "user", "id1": "alice", "t2": "rel", "id2": "knows", "t3": "user", "id3": "bob"}
	m2 := map[string]any{"t1": "user", "id1": "janice", "t2": "rel", "id2": "knows", "t3": "admin", "id3": "danny"}
	m3 := map[string]any{"t1": "admin", "id1": "danny", "t2": "rel", "id2": "knows", "t3": "user", "id3": "alice"}

	s1 := []any{m3, m1, m2}

	testCases := []struct {
		name      string
		data      any
		obj       *RelationMapping
		wantCount int
		want      map[string]*models.Relation
	}{
		{
			name: "ID from value",
			data: "my_key_2",
			obj:  &RelationMapping{Base: "first", SubjectTypeValue: "t1", SubjectIDValue: "id1", PredicateTypeValue: "t2", PredicateIDValue: "id2", ObjectTypeValue: "t3", ObjectIDValue: "id3"},
			want: map[string]*models.Relation{
				"t1::id1|t2::id2|t3::id3": models.NewRelation(
					models.NewEntity("t1", "id1", models.NewAttributeSet()),
					models.NewEntity("t2", "id2", models.NewAttributeSet()),
					models.NewEntity("t3", "id3", models.NewAttributeSet()),
				),
			},
		},
		{
			name: "map (1)",
			data: m1,
			obj:  &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			want: map[string]*models.Relation{
				"user::alice|rel::knows|user::bob": models.NewRelation(
					models.NewEntity("user", "alice", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("user", "bob", models.NewAttributeSet()),
				),
			},
		},
		{
			name: "map (2)",
			data: m2,
			obj:  &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			want: map[string]*models.Relation{
				"user::janice|rel::knows|admin::danny": models.NewRelation(
					models.NewEntity("user", "janice", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("admin", "danny", models.NewAttributeSet()),
				),
			},
		},
		{
			name: "slice",
			data: s1,
			obj:  &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			want: map[string]*models.Relation{
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
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ent := models.NewEntitySet()
			rel := models.NewRelationSet(ent)

			r := &runner{logger: logger, manager: &manager{logger: logger, addEntity: ent.AddEntity, addRelation: rel.AddRelation}}
			r.decodeRelationData(tc.data, tc.obj)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				for k := range tc.want {
					want, got := tc.want[k], rel.GetRelation(k)
					require.NotNil(t, got)
					assert.Equal(t, want.UID(), got.UID())
					assert.EqualValues(t, want.Subject(), got.Subject())
					assert.EqualValues(t, want.Predicate(), got.Predicate())
					assert.EqualValues(t, want.Object(), got.Object())
				}
			}
		})
	}
}

func TestDecodeRelation(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{"hello": "world"}
	m2 := map[string]any{"t1": "user", "id1": "alice", "t2": "rel", "id2": "knows", "t3": "user", "id3": "bob"}
	m3 := map[string]any{"t1": "user", "id1": "janice", "t2": "rel", "id2": "knows", "t3": "admin", "id3": "danny"}
	m4 := map[string]any{"t1": "admin", "id1": "danny", "t2": "rel", "id2": "knows", "t3": "user", "id3": "alice"}

	s1 := []any{m4, m2, m3}

	mm2 := map[string]any{"first": m2, "second": m1}
	mm3 := map[string]any{"first": m3, "second": m3}
	mm4 := map[string]any{"first": s1, "second": m4}

	testCases := []struct {
		name      string
		data      any
		obj       *RelationMapping
		wantCount int
		want      map[string]*models.Relation
	}{
		{
			name: "not map and not slice",
			data: map[string]any{"first": 987654321},
			obj:  &RelationMapping{Base: "first", SubjectTypeValue: "t1", SubjectIDValue: "id1", PredicateTypeValue: "t2", PredicateIDValue: "id2", ObjectTypeValue: "t3", ObjectIDValue: "id3"},
			want: map[string]*models.Relation{
				"t1::id1|t2::id2|t3::id3": models.NewRelation(
					models.NewEntity("t1", "id1", models.NewAttributeSet()),
					models.NewEntity("t2", "id2", models.NewAttributeSet()),
					models.NewEntity("t3", "id3", models.NewAttributeSet()),
				),
			},
		},
		{
			name: "map (1)",
			data: mm2,
			obj:  &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			want: map[string]*models.Relation{
				"user::alice|rel::knows|user::bob": models.NewRelation(
					models.NewEntity("user", "alice", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("user", "bob", models.NewAttributeSet()),
				),
			},
		},
		{
			name: "map (2)",
			data: mm3,
			obj:  &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			want: map[string]*models.Relation{
				"user::janice|rel::knows|admin::danny": models.NewRelation(
					models.NewEntity("user", "janice", models.NewAttributeSet()),
					models.NewEntity("rel", "knows", models.NewAttributeSet()),
					models.NewEntity("admin", "danny", models.NewAttributeSet()),
				),
			},
		},
		{
			name: "slice",
			data: mm4,
			obj:  &RelationMapping{Base: "first", SubjectTypeField: "t1", SubjectIDField: "id1", PredicateTypeField: "t2", PredicateIDField: "id2", ObjectTypeField: "t3", ObjectIDField: "id3"},
			want: map[string]*models.Relation{
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
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ent := models.NewEntitySet()
			rel := models.NewRelationSet(ent)

			r := &runner{logger: logger, data: tc.data, manager: &manager{logger: logger, addEntity: ent.AddEntity, addRelation: rel.AddRelation}}
			r.decodeRelation(tc.obj)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				for k := range tc.want {
					want, got := tc.want[k], rel.GetRelation(k)
					require.NotNil(t, got)
					assert.Equal(t, want.UID(), got.UID())
					assert.EqualValues(t, want.Subject(), got.Subject())
					assert.EqualValues(t, want.Predicate(), got.Predicate())
					assert.EqualValues(t, want.Object(), got.Object())
				}
			}
		})
	}
}
