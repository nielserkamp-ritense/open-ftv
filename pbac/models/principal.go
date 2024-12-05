package models

import (
	"fmt"
	"strings"
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
	if principal, ok := a.GetAttribute(AttrPrincipal).(string); ok && strings.Contains(principal, "::") {
		parts := strings.Split(principal, "::")
		return parts[0], parts[1]
	}

	if zaak, ok := a.GetAttribute(AttrZaakType).(string); ok && zaak != "" {
		if taak, ok2 := a.GetAttribute(AttrTaak).(string); ok2 && taak != "" {
			zaak = fmt.Sprintf("%s-%s", zaak, taak)
		}
		return PrincipalZaak, zaak
	}

	if doel, ok := a.GetAttribute(AttrDoelbinding).(string); ok && doel != "" {
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

	if apikey, ok := a.GetAttribute(AttrApiKey).(string); ok && apikey != "" {
		return PrincipalApp, apikey
	}

	return PrincipalInvalid, "invalid"
}
