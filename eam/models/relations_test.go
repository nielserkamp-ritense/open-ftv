package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	entSet1 = NewEntitySet(
		NewEntity("user", "alice", nil),
		NewEntity("user", "bob", nil),
		NewEntity("action", "POST", nil),
		NewEntity("action", "GET", nil),
		NewEntity("resource", "http://localhost/person", nil),
		NewEntity("resource", "http://localhost/search", nil),
	)

	r1 = NewRelationFromUID("user::bob", "x::y", "resource::http://localhost/person", entSet1)
	r2 = NewRelationFromUID("user::alice", "action::POST", "resource::http://localhost/search", entSet1)
	r3 = NewRelationFromUID("user::bob", "action::GET", "resource::http://localhost/person", entSet1)
	r4 = NewRelationFromUID("a::b", "x::y", "resource::http://localhost/search", entSet1)

	relSet1 = NewRelationSet(entSet1, r1, r2, r3, r4)
	relSet2 = NewRelationSet(entSet1, r1, r3)
)

func TestNewRelationSet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      []any
		wantUID []string
	}{
		{
			name: "empty",
		},
		{
			name:    "single",
			in:      []any{r1},
			wantUID: []string{"user::bob|?|resource::http://localhost/person"},
		},
		{
			name: "few",
			in:   []any{r1, r2, r3, r4},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"user::bob|action::GET|resource::http://localhost/person",
				"?|?|resource::http://localhost/search",
			},
		},
		{
			name: "set",
			in:   []any{relSet1},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"user::bob|action::GET|resource::http://localhost/person",
				"?|?|resource::http://localhost/search",
			},
		},
		{
			name: "mixed",
			in:   []any{r4, relSet2, r2},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"user::bob|action::GET|resource::http://localhost/person",
				"?|?|resource::http://localhost/search",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewRelationSet(entSet1, tc.in...)
			require.NotNil(t, got)

			got.IterateRelations(func(r *Relation) {
				assert.Contains(t, tc.wantUID, r.UID())
			})

			for _, uid := range tc.wantUID {
				r := got.GetRelation(uid)
				assert.NotNil(t, r)
			}
		})
	}
}

func TestRelationSet_AddRelation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      []any
		wantUID []string
	}{
		{
			name: "empty",
		},
		{
			name:    "single",
			in:      []any{r1},
			wantUID: []string{"user::bob|?|resource::http://localhost/person"},
		},
		{
			name: "few",
			in:   []any{r1, r2, r3, r4},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"user::bob|action::GET|resource::http://localhost/person",
				"?|?|resource::http://localhost/search",
			},
		},
		{
			name: "set",
			in:   []any{relSet1},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"user::bob|action::GET|resource::http://localhost/person",
				"?|?|resource::http://localhost/search",
			},
		},
		{
			name: "mixed",
			in:   []any{r4, relSet2, r2},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"user::bob|action::GET|resource::http://localhost/person",
				"?|?|resource::http://localhost/search",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewRelationSet(entSet1)
			require.NotNil(t, got)

			for _, in := range tc.in {
				switch q := in.(type) {
				case *Relation:
					got.AddRelation(q)
				case *RelationSet:
					got.MergeRelations(q)
				}
			}

			got.IterateRelations(func(r *Relation) {
				assert.Contains(t, tc.wantUID, r.UID())
			})

			for _, uid := range tc.wantUID {
				r := got.GetRelation(uid)
				assert.NotNil(t, r)
			}
		})
	}
}

func TestRelationSet_AddRelationFromUID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		subject   []string
		predicate []string
		object    []string
		wantUID   []string
	}{
		{
			name: "empty",
		},
		{
			name:      "single",
			subject:   []string{"user::bob"},
			predicate: []string{"x:y"},
			object:    []string{"resource::http://localhost/person"},
			wantUID:   []string{"user::bob|?|resource::http://localhost/person"},
		},
		{
			name:      "few",
			subject:   []string{"user::bob", "user::alice"},
			predicate: []string{"x:y", "action::POST"},
			object:    []string{"resource::http://localhost/person", "resource::http://localhost/search"},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewRelationSet(entSet1)
			require.NotNil(t, got)

			for i := range tc.subject {
				got.AddRelationFromUID(tc.subject[i], tc.predicate[i], tc.object[i])
			}

			got.IterateRelations(func(r *Relation) {
				assert.Contains(t, tc.wantUID, r.UID())
			})

			for _, uid := range tc.wantUID {
				r := got.GetRelation(uid)
				assert.NotNil(t, r)
			}
		})
	}
}

func TestRelationSet_RemoveRelation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		remove  []string
		wantUID []string
	}{
		{
			name: "none",
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"user::bob|action::GET|resource::http://localhost/person",
				"?|?|resource::http://localhost/search",
			},
		},
		{
			name:   "not found",
			remove: []string{"x::y", "a::b"},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"user::bob|action::GET|resource::http://localhost/person",
				"?|?|resource::http://localhost/search",
			},
		},
		{
			name:   "one",
			remove: []string{"user::bob|action::GET|resource::http://localhost/person"},
			wantUID: []string{
				"user::bob|?|resource::http://localhost/person",
				"user::alice|action::POST|resource::http://localhost/search",
				"?|?|resource::http://localhost/search",
			},
		},
		{
			name: "few",
			remove: []string{
				"?|?|resource::http://localhost/search",
				"a::b",
				"user::bob|action::GET|resource::http://localhost/person",
				"user::bob|?|resource::http://localhost/person",
			},
			wantUID: []string{
				"user::alice|action::POST|resource::http://localhost/search",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewRelationSet(entSet1, relSet1)
			require.NotNil(t, got)

			for _, uid := range tc.remove {
				got.RemoveRelation(uid)
			}

			got.IterateRelations(func(r *Relation) {
				assert.Contains(t, tc.wantUID, r.UID())
			})

			for _, uid := range tc.wantUID {
				r := got.GetRelation(uid)
				assert.NotNil(t, r)
			}
		})
	}
}
