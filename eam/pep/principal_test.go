package pep

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func TestDeterminePrincipal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		a     models.AttributeSet
		want1 string
		want2 string
	}{
		{
			name:  "empty",
			a:     models.NewAttributeSet(),
			want1: "invalid",
			want2: "invalid",
		},
		{
			name:  "zaaktype",
			a:     models.NewAttributeSet(models.NewAttribute("zaak_type", "zaak1")),
			want1: "zaak",
			want2: "zaak1",
		},
		{
			name: "zaaktype met taak",
			a: models.NewAttributeSet(
				models.NewAttribute("zaak_type", "zaak2"),
				models.NewAttribute("taak", "controle"),
			),
			want1: "zaak",
			want2: "zaak2-controle",
		},
		{
			name:  "doelbinding",
			a:     models.NewAttributeSet(models.NewAttribute("doelbinding", "subsidie")),
			want1: "doelbinding",
			want2: "subsidie",
		},
		{
			name:  "api-key",
			a:     models.NewAttributeSet(models.NewAttribute("api_key", "12cd45ef")),
			want1: "app",
			want2: "12cd45ef",
		},
		{
			name:  "principal",
			a:     models.NewAttributeSet(models.NewAttribute("principal", "user::bob")),
			want1: "user",
			want2: "bob",
		},
		{
			name:  "rvva",
			a:     models.NewAttributeSet(models.NewAttribute("rvva_id", "7fc3d429-2435-4d5f-864e-62e444fcd906")),
			want1: "activity",
			want2: "7fc3d429-2435-4d5f-864e-62e444fcd906",
		},
		{
			name:  "user",
			a:     models.NewAttributeSet(models.NewAttribute("basic_user", "alice")),
			want1: "user",
			want2: "alice",
		},
		{
			name: "mixed",
			a: models.NewAttributeSet(
				models.NewAttribute("basic_user", "alice"),
				models.NewAttribute("doelbinding", "subsidie"),
				models.NewAttribute("zaak_type", "zaak2"),
				models.NewAttribute("principal", "user::bob"),
				models.NewAttribute("api_key", "12cd45ef"),
				models.NewAttribute("rvva_id", "7fc3d429-2435-4d5f-864e-62e444fcd906"),
				models.NewAttribute("taak", "controle"),
			),
			want1: "activity",
			want2: "7fc3d429-2435-4d5f-864e-62e444fcd906",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p1, p2 := DeterminePrincipal(tc.a)
			assert.Equal(t, tc.want1, string(p1))
			assert.Equal(t, tc.want2, string(p2))
		})
	}
}
