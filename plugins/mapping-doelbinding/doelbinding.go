// Package doelbinding provides request mappers that promote a doelbinding or an
// RvVA-ID to the principal of a PARC.
//
// Both mappers belong to the same "principal substitution from a purpose
// attribute" family: they detect a purpose-of-use identifier in the request
// (either as a context attribute or as a request header) and, when present,
// rewrite the principal to that identifier while preserving the original
// principal under the "client_principal" context attribute.
//
// Importing this package registers the mappers under the configuration names
// "doelbindingtoprincipal" and "rvvatoprincipal".
package doelbinding

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// DoelbindingToPrincipal detects if the *doelbinding* is present, and uses it as the principal.
//
//   - if the principal is already a *doelbinding*, the given parc is returned unmodified.
//   - if the *doelbinding* is present, a new PARC is based on the given PARC with the *doelbinding* as the principal.
//     in this case, the old principal is stored in the context under the key "client_principal".
//   - if the *doelbinding* is not present, the given parc is returned unmodified.
func DoelbindingToPrincipal(parc *models.PARC, opts ...mapping.Option) *models.PARC {
	if parc.Principal.Type() == pep.PrincipalDoelbinding {
		return parc
	}

	var id string
	if id = convert.AnyToString(parc.Context.GetAttributeValue(models.AttrDoelbinding)); id == "" {
		id = mapping.FromHeaders(parc.Context.GetAttributeValue(models.AttrHeaders), []string{models.HeaderDoelbinding}, opts...)
	}

	if id != "" {
		p := &models.PARC{
			Principal: models.NewEntity(pep.PrincipalDoelbinding, id, models.NewAttributeSet()),
			Action:    parc.Action,
			Resource:  parc.Resource,
			Context:   models.NewAttributeSet(parc.Context),
		}

		old := models.EntityToAttribute(parc.Principal)
		p.Context.AddAttribute(models.AttrClientPrincipal, old.Value())
		return p
	}

	return parc
}
