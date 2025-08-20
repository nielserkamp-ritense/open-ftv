package models

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

func TestNewEntity(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		ns       string
		id       string
		attr     *AttributeSet
		parents  []string
		wantUID  string
		wantJSON string
	}{
		{
			name:     "empty",
			wantUID:  "::",
			wantJSON: `{"attributes":[]}`,
		},
		{
			name:     "no attributes, no parents",
			ns:       "entity",
			id:       "x1",
			wantUID:  "entity::x1",
			wantJSON: `{"type":"entity","id":"x1","attributes":[]}`,
		},
		{
			name:     "just attributes",
			ns:       "entity",
			id:       "x2",
			attr:     NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123)),
			wantUID:  "entity::x2",
			wantJSON: `{"type":"entity","id":"x2","attributes":[{"key":"hello","value":"world"},{"key":"int","value":123}]}`,
		},
		{
			name:     "just parents",
			ns:       "entity",
			id:       "x3",
			parents:  []string{"entity::x1", "entity::x2"},
			wantUID:  "entity::x3",
			wantJSON: `{"type":"entity","id":"x3","attributes":[],"parents":["entity::x1","entity::x2"]}`,
		},
		{
			name:     "all",
			ns:       "entity",
			id:       "x4",
			attr:     NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123)),
			parents:  []string{"entity::x3", "entity::x2"},
			wantUID:  "entity::x4",
			wantJSON: `{"type":"entity","id":"x4","attributes":[{"key":"hello","value":"world"},{"key":"int","value":123}],"parents":["entity::x3","entity::x2"]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewEntity(tc.ns, tc.id, tc.attr, tc.parents...)
			require.NotNil(t, got)

			assert.Equal(t, tc.wantUID, got.UID())
			assert.Equal(t, tc.ns, got.Type())
			assert.Equal(t, tc.id, got.ID())
			assert.EqualValues(t, tc.parents, got.Parents())

			if tc.attr != nil {
				assert.True(t, tc.attr.Equals(got.Attributes()))
			} else {
				assert.True(t, NewAttributeSet().Equals(got.Attributes()))
			}

			b, err := got.MarshalJSON()
			require.NoError(t, err)
			require.NotNil(t, b)
			assert.Equal(t, tc.wantJSON, string(b))
		})
	}
}
func TestEntity_WithTitle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		in    *Entity
		title string
	}{
		{
			name:  "no title",
			in:    &Entity{ns: "user", id: "alice", attrs: NewAttributeSet()},
			title: "New title",
		},
		{
			name:  "existing title",
			in:    &Entity{ns: "user", id: "bob", attrs: NewAttributeSet(), title: "Old title"},
			title: "New title 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.WithTitle(tc.title)
			require.NotNil(t, got)
			assert.Equal(t, tc.title, got.Title())
		})
	}
}

func TestEntity_WithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Entity
		desc string
	}{
		{
			name: "no description",
			in:   &Entity{ns: "user", id: "alice", attrs: NewAttributeSet()},
			desc: "New description",
		},
		{
			name: "existing description",
			in:   &Entity{ns: "user", id: "bob", attrs: NewAttributeSet(), description: "Old description"},
			desc: "New description 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.WithDescription(tc.desc)
			require.NotNil(t, got)
			assert.Equal(t, tc.desc, got.Description())
		})
	}
}

func TestEntity_WithTags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		ns   string
		id   string
		tags []string
	}{
		{name: "single", ns: "user", id: "bob", tags: []string{"x"}},
		{name: "few", ns: "user", id: "alice", tags: []string{"x", "y", "z"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewEntity(tc.ns, tc.id, nil)
			require.NotNil(t, got)

			got.WithTags(tc.tags...)

			tags := got.Tags()
			slices.Sort(tags)
			assert.EqualValues(t, tc.tags, tags)

			for i := range tc.tags {
				assert.True(t, got.HasTag(tc.tags[i]))
			}

			assert.False(t, got.HasTag("qqq"))
		})
	}
}

func TestEntity_WithAudit(t *testing.T) {
	t.Parallel()

	now1 := time.Now().UTC()
	now2 := now1.Add(-13 * time.Hour)

	testCases := []struct {
		name  string
		tp    string
		id    string
		time1 time.Time
		user1 string
		time2 time.Time
		user2 string
	}{
		{name: "created", tp: "user", id: "jimmy", user1: "bob", time1: now1},
		{name: "updated", tp: "action", id: "read", user2: "charlie", time2: now1},
		{name: "both", tp: "resource", id: "book1", user1: "alice", time1: now2, user2: "charlie", time2: now1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewEntity(tc.tp, tc.id, NewAttributeSet())
			require.NotNil(t, got)

			got.WithAudit(tc.time1, tc.user1, tc.time2, tc.user2)
			assert.EqualValues(t, tc.time1, got.Created())
			assert.EqualValues(t, tc.user1, got.CreatedBy())
			assert.EqualValues(t, tc.time2, got.Updated())
			assert.EqualValues(t, tc.user2, got.UpdatedBy())
		})
	}
}

func TestEntityToAttribute(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		e         *Entity
		wantKey   string
		wantValue map[string]any
	}{
		{
			name:      "simple",
			e:         NewEntity("type", "id", NewAttributeSet()),
			wantKey:   "type::id",
			wantValue: map[string]any{"type": "type", "id": "id"},
		},
		{
			name:    "with attributes",
			e:       NewEntity("service", "http://localhost", NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123))),
			wantKey: "service::http://localhost",
			wantValue: map[string]any{
				"type": "service",
				"id":   "http://localhost",
				"attributes": map[string]any{
					"hello": "world",
					"int":   123,
				},
			},
		},
		{
			name:      "with parents",
			e:         NewEntity("user", "alice", NewAttributeSet(), "admin::bob"),
			wantKey:   "user::alice",
			wantValue: map[string]any{"type": "user", "id": "alice", "parents": []string{"admin::bob"}},
		},
		{
			name:      "with tags",
			e:         NewEntity("user", "alice", NewAttributeSet()).WithTags("x", "y"),
			wantKey:   "user::alice",
			wantValue: map[string]any{"type": "user", "id": "alice"},
		},
		{
			name: "with all",
			e: NewEntity(
				"user",
				"alice",
				NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 456)),
				"admin::bob",
			).WithTags("q", "z"),
			wantKey: "user::alice",
			wantValue: map[string]any{
				"type": "user",
				"id":   "alice",
				"attributes": map[string]any{
					"int":   456,
					"hello": "world",
				},
				"parents": []string{"admin::bob"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := EntityToAttribute(tc.e)
			assert.Equal(t, tc.wantKey, got.Key())
			assert.EqualValues(t, tc.wantValue, got.Value())
		})
	}
}

func TestEntity_MarshallYAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Entity
		want string
	}{
		{
			name: "empty",
			in:   &Entity{},
			want: "{}\n",
		},
		{
			name: "simple",
			in:   NewEntity("user", "bob", nil),
			want: "type: user\nid: bob\nattributes: {}\n",
		},
		{
			name: "full",
			in: NewEntity(
				"user",
				"bob",
				NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123)),
			).WithTitle("user bob").
				WithDescription("entity for user bob").
				WithTags("x", "y"),
			want: "type: user\nid: bob\ntitle: user bob\ndescription: entity for user bob\nattributes: {}\ntags:\n- x\n- \"y\"\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.in.MarshalYAML()
			require.NoError(t, err)
			assert.Equal(t, tc.want, string(got))
		})
	}
}

func TestEntityFromOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *attributes.Entity
		want *Entity
	}{
		{
			name: "empty",
			in:   &attributes.Entity{},
			want: NewEntity("", "", nil),
		},
		{
			name: "simple",
			in:   &attributes.Entity{Type: "user", Id: "bob", Attributes: []attributes.Attribute{{Key: "hello", Value: "world"}}},
			want: NewEntity("user", "bob", NewAttributeSet(NewAttribute("hello", "world"))),
		},
		{
			name: "full",
			in: &attributes.Entity{
				Type: "user",
				Id:   "bob",
				Attributes: []attributes.Attribute{
					{Key: "hello", Value: "world"},
					{Key: "int", Value: 123, Type: "integer"},
					{Key: "float", Value: 3.14, Type: "xsd:float"},
				},
				Metadata: attributes.Metadata{
					Title:       "user bob",
					Description: "this is the details for bob",
					Tags:        []string{"x", "y"},
				},
			},
			want: NewEntity(
				"user",
				"bob",
				NewAttributeSet(
					NewAttribute("hello", "world"),
					NewAttributeWithType("int", 123, "integer"),
					NewAttributeWithType("float", 3.14, "xsd:float"),
				),
			).WithTitle("user bob").WithDescription("this is the details for bob").WithTags("x", "y"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := EntityFromOAS(tc.in)
			require.NotNil(t, got)
			assert.True(t, tc.want.Equals(got))
		})
	}
}

func TestEntity_ToOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Entity
		want *attributes.Entity
	}{
		{
			name: "empty",
			in:   NewEntity("", "", nil),
			want: &attributes.Entity{Attributes: []attributes.Attribute{}, Metadata: attributes.Metadata{Tags: make([]string, 0)}},
		},
		{
			name: "simple",
			in: NewEntity(
				"user",
				"bob",
				NewAttributeSet(
					NewAttribute("hello", "world"),
					NewAttribute("int", 123),
				),
			),
			want: &attributes.Entity{
				Type: "user",
				Id:   "bob",
				Attributes: []attributes.Attribute{
					{Key: "hello", Value: "world", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
					{Key: "int", Value: 123, Metadata: attributes.Metadata{Tags: []string{}}},
				},
				Metadata: attributes.Metadata{Tags: make([]string, 0)},
			},
		},
		{
			name: "full",
			in: NewEntity(
				"user",
				"bob",
				NewAttributeSet(
					NewAttribute("hello", "world"),
					NewAttribute("int", 123),
				),
			).WithTitle("user bob").WithDescription("this is the details for bob").WithTags("x", "y"),
			want: &attributes.Entity{
				Type: "user",
				Id:   "bob",
				Attributes: []attributes.Attribute{
					{Key: "hello", Value: "world", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
					{Key: "int", Value: 123, Metadata: attributes.Metadata{Tags: []string{}}},
				},
				Metadata: attributes.Metadata{
					Title:       "user bob",
					Description: "this is the details for bob",
					Tags:        []string{"x", "y"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.ToOAS()
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestEntity_ToBundle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Entity
		want *attributes.Entity
	}{
		{
			name: "empty",
			in:   NewEntity("", "", nil),
			want: &attributes.Entity{Attributes: []attributes.Attribute{}},
		},
		{
			name: "simple",
			in: NewEntity(
				"user",
				"bob",
				NewAttributeSet(
					NewAttribute("hello", "world"),
					NewAttribute("int", 123),
				),
			),
			want: &attributes.Entity{
				Type: "user",
				Id:   "bob",
				Attributes: []attributes.Attribute{
					{Key: "hello", Value: "world", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
					{Key: "int", Value: 123, Metadata: attributes.Metadata{Tags: []string{}}},
				},
			},
		},
		{
			name: "full",
			in: NewEntity(
				"user",
				"bob",
				NewAttributeSet(
					NewAttribute("hello", "world"),
					NewAttribute("int", 123),
				),
			).WithTitle("user bob").WithDescription("this is the details for bob").WithTags("x", "y"),
			want: &attributes.Entity{
				Type: "user",
				Id:   "bob",
				Attributes: []attributes.Attribute{
					{Key: "hello", Value: "world", Type: "string", Metadata: attributes.Metadata{Tags: []string{}}},
					{Key: "int", Value: 123, Metadata: attributes.Metadata{Tags: []string{}}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.ToBundle()
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}
