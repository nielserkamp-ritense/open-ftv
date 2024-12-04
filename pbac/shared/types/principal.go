package types

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

// List of supported principal types.
const (
	PrincipalApp         = "app"
	PrincipalDoelbinding = "doelbinding"
	PrincipalInvalid     = "invalid"
	PrincipalJWT         = "jwt"
	PrincipalZaak        = "zaak"
)

// DeterminePrincipal determines the type of principal and its primary key.
func DeterminePrincipal(a AttributeSet) (string, string) {
	if principal, ok := a.GetAttribute(standards.AttrPrincipal).(string); ok && strings.Contains(principal, "::") {
		parts := strings.Split(principal, "::")
		return parts[0], parts[1]
	}

	if zaak, ok := a.GetAttribute(standards.AttrZaakType).(string); ok && zaak != "" {
		if taak, ok2 := a.GetAttribute(standards.AttrTaak).(string); ok2 && taak != "" {
			zaak = fmt.Sprintf("%s-%s", zaak, taak)
		}
		return PrincipalZaak, zaak
	}

	if doel, ok := a.GetAttribute(standards.AttrDoelbinding).(string); ok && doel != "" {
		return PrincipalDoelbinding, doel
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

	if apikey, ok := a.GetAttribute(standards.AttrApiKey).(string); ok && apikey != "" {
		return PrincipalApp, apikey
	}

	return PrincipalInvalid, "invalid"
}
