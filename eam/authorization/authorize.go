package authorization

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// ErrForbidden reports that the caller was not permitted the action on the resource.
var ErrForbidden = errors.New("forbidden")

// Authorizer is the management plane's authorization question: may this caller perform this
// action on this resource? Implemented by the authorizer; declared separately so a service
// depends on the question alone.
type Authorizer interface {
	Authorize(ctx context.Context, caller *RequestPrincipal, action string, resource *models.Entity) error
}

// Authorize decides once whether caller may perform action on resource, and records the
// caller when it may (ADR 0004). It returns nil when permitted, ErrForbidden when not, and
// ErrPrincipalNotRecorded when a permitted non-read caller could not be recorded.
//
// The decision is the role table given with WithPermissions: without one every caller is
// denied (fail closed), unless the authorizer runs with NoAuth. The resource is not yet part
// of the decision; it is carried so the question keeps its shape for the PDP (ADR 0006).
func (a *auth) Authorize(_ context.Context, caller *RequestPrincipal, action string, resource *models.Entity) error {
	if resource == nil {
		return fmt.Errorf("authorize: resource is required")
	}

	if caller == nil {
		caller = SystemRequestPrincipal()
	}

	uid := uuid.New()

	if !a.noAuth && !a.permissions.Allows(caller, action) {
		a.log.Warn("authorization denied", "uid", uid.String(), "principal", caller.Label(), "roles", caller.Roles, "action", action, "resource", resource.UID())
		return ErrForbidden
	}

	if a.debug {
		a.log.Debug("authorization granted", "uid", uid.String(), "principal", caller.Label(), "action", action, "resource", resource.UID())
	}

	return a.recordPrincipal(action == models.ActionRead, &caller.Principal)
}
