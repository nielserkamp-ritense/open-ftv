package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRelationFromUID(t *testing.T) {
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
