package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventType(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   EventType
		want string
	}{
		{name: "empty", want: "<invalid>"},
		{name: "policy added", in: PolicyAdded, want: "policy added"},
		{name: "policy replaced", in: PolicyReplaced, want: "policy replaced"},
		{name: "policy removed", in: PolicyRemoved, want: "policy removed"},
		{name: "attribute added", in: AttributeAdded, want: "attribute added"},
		{name: "attribute replaced", in: AttributeReplaced, want: "attribute replaced"},
		{name: "attribute removed", in: AttributeRemoved, want: "attribute removed"},
		{name: "entity added", in: EntityAdded, want: "entity added"},
		{name: "entity replaced", in: EntityReplaced, want: "entity replaced"},
		{name: "entity removed", in: EntityRemoved, want: "entity removed"},
		{name: "relation added", in: RelationAdded, want: "relation added"},
		{name: "relation replaced", in: RelationReplaced, want: "relation replaced"},
		{name: "relation removed", in: RelationRemoved, want: "relation removed"},
		{name: "invalid", in: 99, want: "<invalid>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.String()
			assert.Equal(t, tc.want, got)
		})
	}
}
