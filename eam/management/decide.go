package management

import (
	"context"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// decide asks the PDP whether caller may perform action on resource; without an authorizer
// every operation is permitted, for deployments and tests that run without authorization.
func decide(ctx context.Context, authorizer authorization.Authorizer, caller *authorization.RequestPrincipal, action string, resource *models.Entity) error {
	if authorizer == nil {
		return nil
	}

	return authorizer.Authorize(ctx, caller, action, resource)
}
