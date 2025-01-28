package network

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestProcessEntity(t *testing.T) {
	testCases := []struct {
		name      string
		tp        string
		id        string
		attrs     models.AttributeSet
		parents   []string
		obj       *EntityObject
		wantCount int
		wantKey   string
		wantValue models.Entity
	}{
		{
			name:      "no type",
			id:        "oops",
			obj:       &EntityObject{Base: "first"},
			wantCount: 1,
		},
		{
			name:      "no id",
			tp:        "oops",
			obj:       &EntityObject{Base: "first"},
			wantCount: 1,
		},
		{
			name:      "all filled",
			tp:        "third",
			id:        "999",
			obj:       &EntityObject{Base: "first"},
			wantKey:   "third::999",
			wantValue: models.NewEntity("third", "999", models.NewAttributeSet()),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ent := models.NewEntitySet()

			r := &runner{logger: logger, manager: &manager{logger: logger, entities: ent, newAttributes: models.NewAttributeSet}}
			r.processEntity(tc.tp, tc.id, tc.attrs, tc.parents, tc.obj)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				got := ent.GetEntity(tc.wantKey)
				require.NotNil(t, got)
				assert.EqualValues(t, tc.wantValue, got)
			}
		})
	}
}

func TestDecodeEntityMap(t *testing.T) {
	a1 := make([]any, 0)
	a2 := map[string]any{"key": "code", "value": 1, "type": "xsd:string"}
	a3 := []any{
		map[string]any{"key": "code", "value": 1, "type": "xsd:integer"},
		map[string]any{"key": "name", "value": "code 1", "type": "xsd:string"},
		map[string]any{"key": "description", "value": "this is code 1", "type": "xsd:string"},
	}

	m1 := map[string]any{"hello": "world", "int": 123, "bool": true, "id": "my_id"}
	m2 := map[string]any{"type": "service", "id": "this_id2", "parents": []string{"a", "b", "c"}}
	m3 := map[string]any{"type": "service", "id": "this_id3", "parents": "type::id"}
	m4 := map[string]any{"type": "service", "id": "this_id4", "parents": []any{"a", "123", true}}
	m5 := map[string]any{"type": "service", "id": "this_id5", "parents": 123.99}
	m6 := map[string]any{"type": "service", "id": "this_id6", "attr": a1}
	m7 := map[string]any{"type": "service", "id": "this_id7", "attr": a2}
	m8 := map[string]any{"type": "service", "id": "this_id8", "attr": a3}

	attr1 := &AttributeObject{Base: "attr", KeyCode: "key", ValueCode: "value", TypeCode: "type"}

	attrMap := map[string]*AttributeObject{
		attr1.Base: attr1,
	}

	testCases := []struct {
		name      string
		m         map[string]any
		obj       *EntityObject
		wantCount int
		wantKey   string
		wantValue models.Entity
	}{
		{
			name:      "no type code, no type value",
			m:         m1,
			obj:       &EntityObject{Base: "first", IdCode: "id"},
			wantCount: 1,
		},
		{
			name:      "invalid type code, no type value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeCode: "abc", IdCode: "id"},
			wantCount: 1,
		},
		{
			name:      "no type code, type value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeValue: "hello", IdCode: "id"},
			wantKey:   "hello::my_id",
			wantValue: models.NewEntity("hello", "my_id", models.NewAttributeSet()),
		},
		{
			name:      "invalid type code, type value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeCode: "abc", TypeValue: "good", IdCode: "int"},
			wantKey:   "good::123",
			wantValue: models.NewEntity("good", "123", models.NewAttributeSet()),
		},
		{
			name:      "valid type code, type value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeCode: "hello", TypeValue: "good", IdCode: "int"},
			wantKey:   "world::123",
			wantValue: models.NewEntity("world", "123", models.NewAttributeSet()),
		},
		{
			name:      "empty map",
			m:         map[string]any{},
			obj:       &EntityObject{Base: "first", TypeCode: "hello", IdCode: "bool"},
			wantCount: 1,
		},
		{
			name:      "no id code, no id value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeCode: "int"},
			wantCount: 1,
		},
		{
			name:      "no id code, id value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeCode: "hello", IdValue: "key1"},
			wantKey:   "world::key1",
			wantValue: models.NewEntity("world", "key1", models.NewAttributeSet()),
		},
		{
			name:      "invalid id code, no id value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeCode: "int", IdCode: "oops"},
			wantCount: 1,
		},
		{
			name:      "invalid id code, id value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeCode: "int", IdCode: "nope", IdValue: "key1"},
			wantKey:   "123::key1",
			wantValue: models.NewEntity("123", "key1", models.NewAttributeSet()),
		},
		{
			name:      "valid id code, id value",
			m:         m1,
			obj:       &EntityObject{Base: "first", TypeCode: "hello", IdCode: "int", IdValue: "key1"},
			wantKey:   "world::123",
			wantValue: models.NewEntity("world", "123", models.NewAttributeSet()),
		},
		{
			name:      "parents slice of string",
			m:         m2,
			obj:       &EntityObject{Base: "first", TypeCode: "type", IdCode: "id", ParentsCode: "parents"},
			wantKey:   "service::this_id2",
			wantValue: models.NewEntity("service", "this_id2", models.NewAttributeSet(), "a", "b", "c"),
		},
		{
			name:      "parents string",
			m:         m3,
			obj:       &EntityObject{Base: "first", TypeCode: "type", IdCode: "id", ParentsCode: "parents"},
			wantKey:   "service::this_id3",
			wantValue: models.NewEntity("service", "this_id3", models.NewAttributeSet(), "type::id"),
		},
		{
			name:      "parents slice of any",
			m:         m4,
			obj:       &EntityObject{Base: "first", TypeCode: "type", IdCode: "id", ParentsCode: "parents"},
			wantKey:   "service::this_id4",
			wantValue: models.NewEntity("service", "this_id4", models.NewAttributeSet(), "a", "123", "true"),
		},
		{
			name:      "parents float",
			m:         m5,
			obj:       &EntityObject{Base: "first", TypeCode: "type", IdCode: "id", ParentsCode: "parents"},
			wantKey:   "service::this_id5",
			wantValue: models.NewEntity("service", "this_id5", models.NewAttributeSet(), "123.99"),
		},
		{
			name:      "no attributes",
			m:         m6,
			obj:       &EntityObject{Base: "first", TypeCode: "type", IdCode: "id", Attributes: attrMap},
			wantKey:   "service::this_id6",
			wantValue: models.NewEntity("service", "this_id6", models.NewAttributeSet()),
		},
		{
			name:    "one attribute",
			m:       m7,
			obj:     &EntityObject{Base: "first", TypeCode: "type", IdCode: "id", Attributes: attrMap},
			wantKey: "service::this_id7",
			wantValue: models.NewEntity("service", "this_id7", models.NewAttributeSet(
				models.NewOriginalAttribute("code", "1", 1, "xsd:string"),
			)),
		},
		{
			name:    "few attributes",
			m:       m8,
			obj:     &EntityObject{Base: "first", TypeCode: "type", IdCode: "id", Attributes: attrMap},
			wantKey: "service::this_id8",
			wantValue: models.NewEntity("service", "this_id8", models.NewAttributeSet(
				models.NewOriginalAttribute("code", int64(1), 1, "xsd:integer"),
				models.NewOriginalAttribute("name", "code 1", "code 1", "xsd:string"),
				models.NewOriginalAttribute("description", "this is code 1", "this is code 1", "xsd:string"),
			)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ent := models.NewEntitySet()

			r := &runner{logger: logger, manager: &manager{logger: logger, entities: ent, newAttributes: models.NewAttributeSet}}
			r.decodeEntityMap(tc.m, tc.obj)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				got := ent.GetEntity(tc.wantKey)
				require.NotNil(t, got)

				assert.Equal(t, tc.wantValue.UID(), got.UID())
				assert.Equal(t, tc.wantValue.Type(), got.Type())
				assert.Equal(t, tc.wantValue.ID(), got.ID())
				assert.EqualValues(t, tc.wantValue.Parents(), got.Parents())

				tc.wantValue.Attributes().IterateAttributes(func(attr1 models.Attribute) {
					attr2 := got.Attributes().GetAttribute(attr1.Key())
					assert.EqualValues(t, attr1, attr2)
				})

				got.Attributes().IterateAttributes(func(attr1 models.Attribute) {
					attr2 := tc.wantValue.Attributes().GetAttribute(attr1.Key())
					assert.EqualValues(t, attr1, attr2)
				})
			}
		})
	}
}

func TestDecodeEntityData(t *testing.T) {
	m1 := map[string]any{"hello": "world", "int": 123, "bool": true, "type": "xsd:short"}
	m2 := map[string]any{"hello": "mars", "int": 321, "bool": false, "type": "xsd:string"}
	m3 := map[string]any{"hello": "jupiter", "int": true, "bool": 123}

	s1 := []any{m3, m1, m2}

	testCases := []struct {
		name      string
		data      any
		obj       *EntityObject
		wantCount int
		want      map[string]models.Entity
	}{
		{
			name: "ID from value",
			data: "my_key_2",
			obj:  &EntityObject{Base: "first", TypeValue: "key", IdFromValue: true},
			want: map[string]models.Entity{"key::my_key_2": models.NewEntity("key", "my_key_2", models.NewAttributeSet())},
		},
		{
			name: "map (1)",
			data: m1,
			obj:  &EntityObject{Base: "first", TypeCode: "hello", IdCode: "int"},
			want: map[string]models.Entity{"world::123": models.NewEntity("world", "123", models.NewAttributeSet())},
		},
		{
			name: "map (2)",
			data: m2,
			obj:  &EntityObject{Base: "first", TypeCode: "hello", IdCode: "bool"},
			want: map[string]models.Entity{"mars::false": models.NewEntity("mars", "false", models.NewAttributeSet())},
		},
		{
			name: "slice",
			data: s1,
			obj:  &EntityObject{Base: "first", TypeCode: "hello", IdCode: "int"},
			want: map[string]models.Entity{
				"world::123":    models.NewEntity("world", "123", models.NewAttributeSet()),
				"mars::321":     models.NewEntity("mars", "321", models.NewAttributeSet()),
				"jupiter::true": models.NewEntity("jupiter", "true", models.NewAttributeSet()),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ent := models.NewEntitySet()

			r := &runner{logger: logger, manager: &manager{logger: logger, entities: ent, newAttributes: models.NewAttributeSet}}
			r.decodeEntityData(tc.data, tc.obj)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				for k := range tc.want {
					want, got := tc.want[k], ent.GetEntity(k)
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
			}
		})
	}
}

func TestDecodeEntity(t *testing.T) {
	m1 := map[string]interface{}{"hello": "world"}
	m2 := map[string]any{"hello": "world", "int": 123, "bool": true, "type": "xsd:short"}
	m3 := map[string]any{"hello": "mars", "int": 321, "bool": false, "type": "xsd:string"}
	m4 := map[string]any{"hello": "jupiter", "int": true, "bool": 123}

	s1 := []any{m4, m2, m3}

	mm1 := map[string]any{"first": "this_key", "second": m2}
	mm2 := map[string]any{"first": m2, "second": m1}
	mm3 := map[string]any{"first": m3, "second": m3}
	mm4 := map[string]any{"first": s1, "second": m4}

	testCases := []struct {
		name      string
		data      any
		obj       *EntityObject
		wantCount int
		want      map[string]models.Entity
	}{
		{
			name: "ID from value",
			data: mm1,
			obj:  &EntityObject{Base: "first", TypeValue: "key", IdFromValue: true},
			want: map[string]models.Entity{"key::this_key": models.NewEntity("key", "this_key", models.NewAttributeSet())},
		},
		{
			name: "not map and not slice",
			data: map[string]any{"first": 987654321},
			obj:  &EntityObject{Base: "first", TypeValue: "key"},
			want: map[string]models.Entity{"key::987654321": models.NewEntity("key", "987654321", models.NewAttributeSet())},
		},
		{
			name: "map (1)",
			data: mm2,
			obj:  &EntityObject{Base: "first", TypeCode: "hello", IdCode: "int"},
			want: map[string]models.Entity{"world::123": models.NewEntity("world", "123", models.NewAttributeSet())},
		},
		{
			name: "map (2)",
			data: mm3,
			obj:  &EntityObject{Base: "first", TypeCode: "hello", IdCode: "bool"},
			want: map[string]models.Entity{"mars::false": models.NewEntity("mars", "false", models.NewAttributeSet())},
		},
		{
			name: "slice",
			data: mm4,
			obj:  &EntityObject{Base: "first", TypeCode: "hello", IdCode: "int"},
			want: map[string]models.Entity{
				"world::123":    models.NewEntity("world", "123", models.NewAttributeSet()),
				"mars::321":     models.NewEntity("mars", "321", models.NewAttributeSet()),
				"jupiter::true": models.NewEntity("jupiter", "true", models.NewAttributeSet()),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ent := models.NewEntitySet()

			r := &runner{logger: logger, data: tc.data, manager: &manager{logger: logger, entities: ent, newAttributes: models.NewAttributeSet}}
			r.decodeEntity(tc.obj)

			if tc.wantCount > 0 {
				assert.Equal(t, tc.wantCount, h.Count())
			} else {
				require.Zero(t, h.Count())

				for k := range tc.want {
					want, got := tc.want[k], ent.GetEntity(k)
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
			}
		})
	}
}
