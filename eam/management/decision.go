package management

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authlog"
)

// DecisionService serves the entries of the Authorization Decision Log (ADL). The log is
// written by the PDPs and only ever searched here: one decision followed by one search.
type DecisionService struct {
	searcher   search.Searcher
	authorizer authorization.Authorizer
}

// NewDecisionService instantiates the decision service. A nil authorizer permits everything,
// for deployments and tests that run without authorization.
func NewDecisionService(searcher search.Searcher, authorizer authorization.Authorizer) *DecisionService {
	return &DecisionService{searcher: searcher, authorizer: authorizer}
}

// Search returns the log entries matching the criteria. Every search is a read of the
// collection, also when the criteria name one entry: which entries a caller may see is
// filtering, not a decision (ADR 0006).
func (s *DecisionService) Search(ctx context.Context, caller *authorization.RequestPrincipal, c *search.Criteria) (oas.AuthlogEntries, error) {
	if err := s.DecideSearch(ctx, caller); err != nil {
		return nil, err
	}

	return s.searcher.Search(ctx, c)
}

// DecideSearch asks whether caller may search the log, without searching: it lets a handler
// answer malformed criteria in decision order, the denial before the parse error.
func (s *DecisionService) DecideSearch(ctx context.Context, caller *authorization.RequestPrincipal) error {
	return decide(ctx, s.authorizer, caller, models.ActionRead, decisionCollection())
}

// decisionCollection is the resource of a collection read.
func decisionCollection() *models.Entity {
	return models.NewEntity(models.EntityTypeDecision, models.ResourceCollection, models.NewAttributeSet())
}
