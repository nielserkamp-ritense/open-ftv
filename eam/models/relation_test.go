package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRelationFromUID(t *testing.T) {
	t.Parallel()

	set := NewEntitySet(
		NewEntity("user", "alice", nil),
		NewEntity("user", "bob", nil),
		NewEntity("action", "POST", nil),
		NewEntity("action", "GET", nil),
		NewEntity("resource", "http://localhost/person", nil),
		NewEntity("resource", "http://localhost/search", nil),
	)

	testCases := []struct {
		name      string
		subject   string
		predicate string
		object    string
		wantUID   string
	}{
		{
			name:      "all unknown",
			subject:   "x::y",
			predicate: "a::b",
			object:    "q::r",
			wantUID:   "?|?|?",
		},
		{
			name:      "unknown subject",
			subject:   "x::y",
			predicate: "action::POST",
			object:    "resource::http://localhost/search",
			wantUID:   "?|action::POST|resource::http://localhost/search",
		},
		{
			name:      "unknown predicate",
			subject:   "user::bob",
			predicate: "a:b",
			object:    "resource::http://localhost/search",
			wantUID:   "user::bob|?|resource::http://localhost/search",
		},
		{
			name:      "unknown object",
			subject:   "user::alice",
			predicate: "action::GET",
			object:    "q:r",
			wantUID:   "user::alice|action::GET|?",
		},
		{
			name:      "all good",
			subject:   "user::bob",
			predicate: "action::GET",
			object:    "resource::http://localhost/person",
			wantUID:   "user::bob|action::GET|resource::http://localhost/person",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewRelationFromUID(tc.subject, tc.predicate, tc.object, set)
			require.NotNil(t, got)

			assert.Equal(t, tc.wantUID, got.UID())

			if s := set.GetEntity(tc.subject); s != nil {
				assert.Equal(t, s, got.Subject())
			}

			if p := set.GetEntity(tc.predicate); p != nil {
				assert.Equal(t, p, got.Predicate())
			}

			if o := set.GetEntity(tc.object); o != nil {
				assert.Equal(t, o, got.Object())
			}
		})
	}
}

func TestRelationToAttribute(t *testing.T) {
	t.Parallel()

	t.Run("relation to attribute", func(t *testing.T) {
		s := NewEntity("user", "alice", nil)
		p := NewEntity("action", "read", nil)
		o := NewEntity("book", "123456", nil)

		r := NewRelation(s, p, o)
		require.NotNil(t, r)

		got := RelationToAttribute(r)
		assert.Equal(t, "user::alice|action::read|book::123456", got.Key())

		m := map[string]any{
			"subject":   map[string]any{"type": "user", "id": "alice"},
			"predicate": map[string]any{"type": "action", "id": "read"},
			"object":    map[string]any{"type": "book", "id": "123456"},
		}
		assert.EqualValues(t, m, got.Value())
	})
}

func TestRelation_WithTitle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		in    *Relation
		title string
	}{
		{
			name: "no title",
			in: &Relation{
				subject:   NewEntity("user", "alice", nil),
				predicate: NewEntity("action", "POST", nil),
				object:    NewEntity("resource", "http://localhost/person", nil),
			},
			title: "New title",
		},
		{
			name: "existing title",
			in: &Relation{
				subject:   NewEntity("user", "alice", nil),
				predicate: NewEntity("action", "POST", nil),
				object:    NewEntity("resource", "http://localhost/person", nil),
				title:     "Old title",
			},
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

func TestRelation_WithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Relation
		desc string
	}{
		{
			name: "no description",
			in: &Relation{
				subject:   NewEntity("user", "alice", nil),
				predicate: NewEntity("action", "POST", nil),
				object:    NewEntity("resource", "http://localhost/person", nil),
			},
			desc: "New description",
		},
		{
			name: "existing description",
			in: &Relation{
				subject:     NewEntity("user", "alice", nil),
				predicate:   NewEntity("action", "POST", nil),
				object:      NewEntity("resource", "http://localhost/person", nil),
				description: "old description",
			},
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

func TestRelation_WithTags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		s    *Entity
		p    *Entity
		o    *Entity
		tags []string
	}{
		{
			name: "single",
			s:    NewEntity("user", "alice", nil),
			p:    NewEntity("action", "POST", nil),
			o:    NewEntity("resource", "http://localhost/person", nil),
			tags: []string{"x"},
		},
		{
			name: "few",
			s:    NewEntity("user", "alice", nil),
			p:    NewEntity("action", "POST", nil),
			o:    NewEntity("resource", "http://localhost/person", nil),
			tags: []string{"x", "y", "z"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewRelation(tc.s, tc.p, tc.o)
			require.NotNil(t, got)

			got2 := got.WithTags(tc.tags...)
			assert.Equal(t, got, got2)
			assert.EqualValues(t, tc.tags, got.Tags())

			for i := range tc.tags {
				assert.True(t, got.HasTag(tc.tags[i]))
			}

			assert.False(t, got.HasTag("qqq"))
		})
	}
}
