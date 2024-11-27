package standards

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/types"
)

// List of supported principal types.
const (
	PrincipalZaak         = "zaak"
	PrincipaleDoelbinding = "doelbinding"
	PrincipalJWT          = "jwt"
	PrincipalApp          = "app"
	PrincipalInvalid      = "invalid"
)

// DeterminePrincipal determines the type of principal and its primary key.
func DeterminePrincipal(a types.AttributeSet) (string, string) {
	if zaak, ok := a.GetAttribute(AttrZaakType).(string); ok && zaak != "" {
		if taak, ok2 := a.GetAttribute(AttrTaak).(string); ok2 && taak != "" {
			zaak = fmt.Sprintf("%s-%s", zaak, taak)
		}
		return PrincipalZaak, zaak
	}

	if doel, ok := a.GetAttribute(AttrDoelbinding).(string); ok && doel != "" {
		return PrincipaleDoelbinding, doel
	}

	// TODO: this is dubious; can we really determine the Principal from a JWT?
	// if jwt := a.GetAttribute(AttrJWT); jwt != nil {
	// 	if valid, ok2 := a.GetAttribute(AttrValid).(cedar.Boolean); ok2 && bool(valid) {
	// 		switch t := a.GetAttribute(AttrClaims).(type) {
	// 		case string:
	// 			return PrincipalJWT, t
	//
	// 		case []string:
	// 			out := bytes.Buffer{}
	// 			for i := range t {
	// 				out.WriteString(t[i])
	// 				out.WriteByte(',')
	// 			}
	// 			out.Truncate(1)
	// 			return PrincipalJWT, out.String()
	// 		}
	// 	}
	// }

	if apikey, ok := a.GetAttribute(AttrApiKey).(string); ok && apikey != "" {
		return PrincipalApp, apikey
	}

	return PrincipalInvalid, "invalid"
}
