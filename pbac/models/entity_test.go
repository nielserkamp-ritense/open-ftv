package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntity(t *testing.T) {
	testCases := []struct {
		name     string
		ns       string
		id       string
		attr     AttributeSet
		parents  []string
		wantUID  string
		wantJSON string
	}{
		{
			name:     "empty",
			wantUID:  "::",
			wantJSON: `{}`,
		},
		{
			name:     "no attributes, no parents",
			ns:       "entity",
			id:       "x1",
			wantUID:  "entity::x1",
			wantJSON: `{"type":"entity","id":"x1"}`,
		},
		{
			name:     "just attributes",
			ns:       "entity",
			id:       "x2",
			attr:     NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123)),
			wantUID:  "entity::x2",
			wantJSON: `{"type":"entity","id":"x2","attributes":{}}`,
		},
		{
			name:     "just parents",
			ns:       "entity",
			id:       "x3",
			parents:  []string{"entity::x1", "entity::x2"},
			wantUID:  "entity::x3",
			wantJSON: `{"type":"entity","id":"x3","parents":["entity::x1","entity::x2"]}`,
		},
		{
			name:     "all",
			ns:       "entity",
			id:       "x4",
			attr:     NewAttributeSet(NewAttribute("hello", "world"), NewAttribute("int", 123)),
			parents:  []string{"entity::x3", "entity::x2"},
			wantUID:  "entity::x4",
			wantJSON: `{"type":"entity","id":"x4","attributes":{},"parents":["entity::x3","entity::x2"]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewEntity(tc.ns, tc.id, tc.attr, tc.parents...)
			require.NotNil(t, got)

			got2, ok := got.(*entity)
			require.True(t, ok)
			require.NotNil(t, got2)

			assert.Equal(t, tc.wantUID, got.UID())
			assert.Equal(t, tc.ns, got.Type())
			assert.Equal(t, tc.id, got.ID())
			assert.Equal(t, tc.attr, got.Attributes())
			assert.Equal(t, tc.parents, got.Parents())

			b, err := got.MarshalJSON()
			require.NoError(t, err)
			require.NotNil(t, b)
			assert.Equal(t, tc.wantJSON, string(b))
		})
	}
}
