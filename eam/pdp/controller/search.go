package controller

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// Search implements the Controller.Search interface.
func (b *Base) Search(uid string, req *models.PARC) ([]string, error) {
	switch {
	case req.Principal.ID() == "":
		return b.searchPrincipal(uid, req)
	case req.Action.ID() == "":
		return b.searchAction(uid, req)
	case req.Resource.ID() == "":
		return b.searchResource(uid, req)
	default:
		return nil, fmt.Errorf("invalid search type")
	}
}

func (b *Base) searchPrincipal(uid string, req *models.PARC) ([]string, error) {
	s := fmt.Sprintf("subjects.%s", req.Principal.Type())
	return b.doSearch(uid, s, func(id string) models.PARC {
		return models.PARC{
			Principal: models.NewEntity(req.Principal.Type(), id, req.Principal.Attributes()),
			Action:    req.Action,
			Resource:  req.Resource,
			Context:   req.Context,
		}
	})
}

func (b *Base) searchAction(uid string, req *models.PARC) ([]string, error) {
	s := fmt.Sprintf("actions.%s", req.Resource.Type())
	return b.doSearch(uid, s, func(id string) models.PARC {
		return models.PARC{
			Principal: req.Principal,
			Action:    models.NewEntity(req.Action.Type(), id, req.Action.Attributes()),
			Resource:  req.Resource,
			Context:   req.Context,
		}
	})
}

func (b *Base) searchResource(uid string, req *models.PARC) ([]string, error) {
	s := fmt.Sprintf("resources.%s", req.Resource.Type())
	return b.doSearch(uid, s, func(id string) models.PARC {
		return models.PARC{
			Principal: req.Principal,
			Action:    req.Action,
			Resource:  models.NewEntity(req.Resource.Type(), id, req.Resource.Attributes()),
			Context:   req.Context,
		}
	})
}

func (b *Base) doSearch(uid, s string, f func(id string) models.PARC) ([]string, error) {
	v := convert.AnyToString(b.PIP.GetAttributeValue(s))
	if v == "" {
		return nil, fmt.Errorf("attribute [%s] not found", s)
	}

	list := strings.Split(convert.AnyToString(v), ",")
	evList := models.Batch{Items: make([]models.PARC, 0, len(list)), Semantics: models.EvaluateAll}

	for i := range list {
		evList.Items = append(evList.Items, f(list[i]))
	}

	resp, err := b.Self.Batch(uid, &evList)
	if err != nil {
		return nil, err
	}

	var out []string
	for i := range resp {
		if resp[i].Allowed {
			out = append(out, list[i])
		}
	}
	return out, nil
}
