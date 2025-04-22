package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pep"
)

func TestRvvaAsPrincipal(t *testing.T) {
	t.Parallel()

	alice := models.NewEntity("user", "alice", models.NewAttributeSet(models.NewAttributeWithType("admin", false, "xsd:boolean")))
	bob := models.NewEntity("user", "bob", models.NewAttributeSet(models.NewAttributeWithType("admin", true, "xsd:boolean")))
	rvva := models.NewEntity(pep.PrincipalRVVA, "id1", models.NewAttributeSet(models.NewAttributeWithType(models.AttrDoelbinding, "kap[vergunning", "xsd:string")))
	action := models.NewEntity("name", "can_read", models.NewAttributeSet(models.NewAttributeWithType("method", "POST", "xsd:string")))
	resource := models.NewEntity("service", "brp-personen", models.NewAttributeSet(models.NewAttributeWithType("owner", "RvIG", "xsd:string")))

	attr1 := models.NewAttributeWithType(models.AttrRvvaID, "xyz", "xsd:string")
	attr2 := models.NewAttribute(models.AttrHeaders, map[string]string{"doelbinding": "xyz", models.HeaderRvvaID: "abc"})
	attr3 := models.NewAttributeWithType(models.AttrRvvaID, 123456, "xsd:integer")

	context1 := models.NewAttributeSet()
	context2 := models.NewAttributeSet(attr1)
	context3 := models.NewAttributeSet(attr2)
	context4 := models.NewAttributeSet(attr2, attr3)

	testCases := []struct {
		name         string
		parc         *models.PARC
		wantType     string
		wantID       string
		wantOriginal models.Attribute
	}{
		{
			name:     "not found",
			parc:     &models.PARC{Principal: alice, Action: action, Resource: resource, Context: context1},
			wantType: alice.Type(),
			wantID:   alice.ID(),
		},
		{
			name:     "already rvva",
			parc:     &models.PARC{Principal: rvva, Action: action, Resource: resource, Context: context1},
			wantType: pep.PrincipalRVVA,
			wantID:   "id1",
		},
		{
			name:         "rvva in context",
			parc:         &models.PARC{Principal: alice, Action: action, Resource: resource, Context: context2},
			wantType:     pep.PrincipalRVVA,
			wantID:       "xyz",
			wantOriginal: models.EntityToAttribute(alice),
		},
		{
			name:         "rvva in headers",
			parc:         &models.PARC{Principal: bob, Action: action, Resource: resource, Context: context3},
			wantType:     pep.PrincipalRVVA,
			wantID:       "abc",
			wantOriginal: models.EntityToAttribute(bob),
		},
		{
			name:         "rvva in context and headers",
			parc:         &models.PARC{Principal: bob, Action: action, Resource: resource, Context: context4},
			wantType:     pep.PrincipalRVVA,
			wantID:       "123456",
			wantOriginal: models.EntityToAttribute(bob),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := RvvaToPrincipal(tc.parc)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantType, got.Principal.Type())
			assert.Equal(t, tc.wantID, got.Principal.ID())

			if tc.wantOriginal != nil {
				orig := got.Context.GetAttributeValue(models.AttrClientPrincipal)
				require.NotNil(t, orig)
				assert.True(t, models.AttributeEqual(tc.wantOriginal, models.NewAttribute(tc.wantOriginal.Key(), orig)))
			}
		})
	}
}
