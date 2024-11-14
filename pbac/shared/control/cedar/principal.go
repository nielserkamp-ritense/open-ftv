package cedar

import (
	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

// DeterminePrincipal determines the type of principal and its primary key.
func DeterminePrincipal(a *attributes) (cedar.EntityType, cedar.String) {
	p1, p2 := standards.DeterminePrincipal(a)
	return cedar.EntityType(p1), cedar.String(p2)
}
