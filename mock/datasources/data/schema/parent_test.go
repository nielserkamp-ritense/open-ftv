package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParent_FQDN(t *testing.T) {
	testCases := []struct {
		name   string
		id     string
		parent *Parent
		want   string
	}{
		{name: "no id, no parent"},
		{name: "id with dots", id: "id.id.id", want: "id.id.id"},
		{name: "no parent", id: "id1", want: "id1"},
		{name: "parent", id: "id2", parent: &Parent{ID: "id1"}, want: "id1.id2"},
		{name: "multiple parents", id: "id3", parent: &Parent{ID: "id1", parent: &Parent{ID: "id2"}}, want: "id2.id1.id3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Parent{ID: tc.id, parent: tc.parent}
			got := p.FQDN()
			assert.Equal(t, tc.want, got)
		})
	}
}
