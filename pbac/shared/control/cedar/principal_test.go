package cedar

import (
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/stretchr/testify/assert"
)

func TestDeterminePrincipal(t *testing.T) {
	testCases := []struct {
		name  string
		a     *attributes
		want1 string
		want2 string
	}{
		{
			name:  "empty",
			a:     &attributes{},
			want1: "Invalid",
			want2: "invalid",
		},
		{
			name: "zaaktype",
			a: &attributes{
				set: cedar.RecordMap{"zaak-type": cedar.String("zaak1")},
			},
			want1: "Zaak",
			want2: "zaak1",
		},
		{
			name: "zaaktype met taak",
			a: &attributes{
				set: cedar.RecordMap{
					"zaak-type": cedar.String("zaak2"),
					"taak":      cedar.String("controle"),
				},
			},
			want1: "Zaak",
			want2: "zaak2-controle",
		},
		{
			name: "doelbinding",
			a: &attributes{
				set: cedar.RecordMap{"doelbinding": cedar.String("subsidie")},
			},
			want1: "Doelbinding",
			want2: "subsidie",
		},
		{
			name: "api-key",
			a: &attributes{
				set: cedar.RecordMap{"api-key": cedar.String("12cd45ef")},
			},
			want1: "App",
			want2: "12cd45ef",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p1, p2 := DeterminePrincipal(tc.a)
			assert.Equal(t, cedar.EntityType(tc.want1), p1)
			assert.Equal(t, cedar.String(tc.want2), p2)
		})
	}
}

var jwtb64 string

func init() {

}
