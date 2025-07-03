package mapping

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// RvvaToPrincipal detects if an RvVA-ID is present, and uses it as the principal.
//
//   - if the principal is already an RvVA-ID, the given parc is returned unmodified.
//   - if the RvVA-ID is present, a new PARC is based on the given PARC with the RvVA-ID as the principal.
//     in this case, the old principal is stored in the context under the key "client_principal".
//   - if the RvVA-ID is not present, the given parc is returned unmodified.
func RvvaToPrincipal(parc *models.PARC, opts ...Option) *models.PARC {
	if parc.Principal.Type() == pep.PrincipalRVVA {
		return parc
	}

	b := &base{headerKeys: []string{models.HeaderRvvaID, models.HeaderObsoleteRvvaID}}
	b.configure(opts)

	var id string
	if id = convert.AnyToString(parc.Context.GetAttributeValue(models.AttrRvvaID)); id == "" {
		id = b.fromHeaders(parc.Context.GetAttributeValue(models.AttrHeaders))
	}

	if id != "" {
		p := &models.PARC{
			Principal: models.NewEntity(pep.PrincipalRVVA, id, models.NewAttributeSet()),
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
