package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func TestBodyToContext(t *testing.T) {
	t.Parallel()

	subject := models.NewEntity("user", "alice", models.NewAttributeSet())
	resource := models.NewEntity("service", "brp-personen", models.NewAttributeSet())
	context0 := models.NewAttributeSet(models.NewAttribute("body", "hello world"))

	body1 := models.NewAttribute("body", "")
	body2 := models.NewAttribute("body", "hello world")
	body3 := models.NewAttribute("body", `{\"hello\":\"world\"}`)
	body4 := models.NewAttribute("body", "<start>hello world</start>")

	action0 := models.NewEntity("name", "POST", models.NewAttributeSet())
	action1 := models.NewEntity("name", "POST", models.NewAttributeSet(body1))
	action2 := models.NewEntity("name", "POST", models.NewAttributeSet(body2))
	action3 := models.NewEntity("name", "POST", models.NewAttributeSet(body3))
	action4 := models.NewEntity("name", "POST", models.NewAttributeSet(body4))

	context3 := models.NewAttributeSet(context0, models.NewAttribute("body", map[string]any{"hello": "world"}))

	nodes := []map[string]any{
		{
			"start": map[string]any{
				"attributes": []map[string]any{},
				"cdata":      "hello world",
				"nodes":      []map[string]any{},
			},
		},
	}
	context4 := models.NewAttributeSet(context0, models.NewAttribute("body", map[string]any{"nodes": nodes}))

	testCases := []struct {
		name string
		in   *models.PARC
		want *models.PARC
	}{
		{
			name: "no body",
			in: &models.PARC{
				Principal: subject,
				Action:    action0,
				Resource:  resource,
				Context:   context0,
			},
			want: &models.PARC{
				Principal: subject,
				Action:    action0,
				Resource:  resource,
				Context:   context0,
			},
		},
		{
			name: "empty body",
			in: &models.PARC{
				Principal: subject,
				Action:    action1,
				Resource:  resource,
				Context:   context0,
			},
			want: &models.PARC{
				Principal: subject,
				Action:    action1,
				Resource:  resource,
				Context:   context0,
			},
		},
		{
			name: "unsupported body",
			in: &models.PARC{
				Principal: subject,
				Action:    action2,
				Resource:  resource,
				Context:   context0,
			},
			want: &models.PARC{
				Principal: subject,
				Action:    action2,
				Resource:  resource,
				Context:   context0,
			},
		},
		{
			name: "valid json",
			in: &models.PARC{
				Principal: subject,
				Action:    action3,
				Resource:  resource,
				Context:   context0,
			},
			want: &models.PARC{
				Principal: subject,
				Action:    action3,
				Resource:  resource,
				Context:   context3,
			},
		},
		{
			name: "valid xml",
			in: &models.PARC{
				Principal: subject,
				Action:    action4,
				Resource:  resource,
				Context:   context0,
			},
			want: &models.PARC{
				Principal: subject,
				Action:    action4,
				Resource:  resource,
				Context:   context4,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := BodyToContext(tc.in)
			require.NotNil(t, got)

			assert.True(t, models.EntityEqual(tc.want.Principal, got.Principal))
			assert.True(t, models.EntityEqual(tc.want.Action, got.Action))
			assert.True(t, models.EntityEqual(tc.want.Resource, got.Resource))
			assert.True(t, models.AttributesEqual(tc.want.Context, got.Context))
		})
	}
}
