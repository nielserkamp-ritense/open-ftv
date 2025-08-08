package mapping

import (
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
func DoelbindingToPrincipal(parc *models.PARC, opts ...Option) *models.PARC {
	if parc.Principal.Type() == pep.PrincipalDoelbinding {
		return parc
	}

	b := &base{headerKeys: []string{models.HeaderDoelbinding}}
	b.configure(opts)

	var id string
	if id = convert.AnyToString(parc.Context.GetAttributeValue(models.AttrDoelbinding)); id == "" {
		id = b.fromHeaders(parc.Context.GetAttributeValue(models.AttrHeaders))
	}

	if id != "" {
		p := &models.PARC{
			Principal: models.NewEntity(pep.PrincipalDoelbinding, id, models.NewAttributeSet()),
			Action:    parc.Action,
			Resource:  parc.Resource,
			Context:   models.NewAttributeSet(parc.Context),
		}

		old := models.EntityToAttribute(parc.Principal)
		p.Context.AddAttributeKV(models.AttrClientPrincipal, old.Value())
		return p
	}

	return parc
}
