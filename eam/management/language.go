package management

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// LanguageStore is what the language service needs from the PAP: the supported policy
// languages, without the rest. *pap.PAP satisfies it.
type LanguageStore interface {
	ListLanguages() (oas.Languages, error)
}

// LanguageService serves the policy languages the store supports. The list is fixed by the
// software, so reading it is the only operation: one decision followed by one store call.
type LanguageService struct {
	store      LanguageStore
	authorizer authorization.Authorizer
}

// NewLanguageService instantiates the language service. A nil authorizer permits everything,
// for deployments and tests that run without authorization.
func NewLanguageService(store LanguageStore, authorizer authorization.Authorizer) *LanguageService {
	return &LanguageService{store: store, authorizer: authorizer}
}

// List returns every supported policy language.
func (s *LanguageService) List(ctx context.Context, caller *authorization.RequestPrincipal) (oas.Languages, error) {
	if err := decide(ctx, s.authorizer, caller, models.ActionRead, languageCollection()); err != nil {
		return nil, err
	}

	return s.store.ListLanguages()
}

// languageCollection is the resource of a collection read.
func languageCollection() *models.Entity {
	return models.NewEntity(models.EntityTypeLanguage, models.ResourceCollection, models.NewAttributeSet())
}
