package pep

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// List of supported principal types.
const (
	PrincipalApp         = "app"
	PrincipalDoelbinding = "doelbinding"
	PrincipalInvalid     = "invalid"
	PrincipalRVVA        = "activity"
	PrincipalUser        = "user"
	PrincipalZaak        = "zaak"
)

// DeterminePrincipal determines the type of principal and its primary key.
func DeterminePrincipal(a models.AttributeSet) (string, string) {
	if rvvaID, ok := a.GetAttributeValue(models.AttrRvvaID).(string); ok {
		return PrincipalRVVA, rvvaID
	}

	if principal, ok := a.GetAttributeValue(models.AttrPrincipal).(string); ok && strings.Contains(principal, "::") {
		parts := strings.Split(principal, "::")
		return parts[0], parts[1]
	}

	if zaak, ok := a.GetAttributeValue(models.AttrZaakType).(string); ok && zaak != "" {
		if taak, ok2 := a.GetAttributeValue(models.AttrTaak).(string); ok2 && taak != "" {
			zaak = fmt.Sprintf("%s-%s", zaak, taak)
		}
		return PrincipalZaak, zaak
	}

	if doel, ok := a.GetAttributeValue(models.AttrDoelbinding).(string); ok && doel != "" {
		return PrincipalDoelbinding, doel
	}

	if apikey, ok := a.GetAttributeValue(models.AttrAPIKey).(string); ok && apikey != "" {
		return PrincipalApp, apikey
	}

	if user, ok := a.GetAttributeValue(models.AttrBasicUser).(string); ok && user != "" {
		return PrincipalUser, user
	}

	return PrincipalInvalid, "invalid"
}

func (c *collector) determinePrincipal() {
	if p := c.parc.Principal; p.Type() != "" || p.ID() != "" {
		return
	}

	pt, pid := DeterminePrincipal(c.parc.Context)
	c.parc.Principal = models.NewEntity(pt, pid, models.NewAttributeSet(c.parc.Principal.Attributes()))
}
