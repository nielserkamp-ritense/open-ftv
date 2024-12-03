package standards

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/types"
)

func TestDeterminePrincipal(t *testing.T) {
	testCases := []struct {
		name  string
		a     types.AttributeSet
		want1 string
		want2 string
	}{
		{
			name:  "empty",
			a:     types.NewAttributeSet(),
			want1: "invalid",
			want2: "invalid",
		},
		{
			name:  "zaaktype",
			a:     types.NewAttributeSet(types.NewAttribute("zaak-type", "zaak1")),
			want1: "zaak",
			want2: "zaak1",
		},
		{
			name: "zaaktype met taak",
			a: types.NewAttributeSet(
				types.NewAttribute("zaak-type", "zaak2"),
				types.NewAttribute("taak", "controle"),
			),
			want1: "zaak",
			want2: "zaak2-controle",
		},
		{
			name:  "doelbinding",
			a:     types.NewAttributeSet(types.NewAttribute("doelbinding", "subsidie")),
			want1: "doelbinding",
			want2: "subsidie",
		},
		{
			name:  "api-key",
			a:     types.NewAttributeSet(types.NewAttribute("api-key", "12cd45ef")),
			want1: "app",
			want2: "12cd45ef",
		},
		{
			name:  "principal",
			a:     types.NewAttributeSet(types.NewAttribute("principal", "user::bob")),
			want1: "user",
			want2: "bob",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p1, p2 := DeterminePrincipal(tc.a)
			assert.Equal(t, tc.want1, string(p1))
			assert.Equal(t, tc.want2, string(p2))
		})
	}
}
