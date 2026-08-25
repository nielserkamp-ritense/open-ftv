package authorization

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// ErrForbidden reports that the PDP did not permit the action on the resource.
var ErrForbidden = errors.New("forbidden")

// Decider is the management plane's authorization question: may this caller perform this
// action on this resource? Implemented by the authorizer; declared separately so a service
// depends on the question alone.
type Decider interface {
	Decide(ctx context.Context, caller *RequestPrincipal, action string, resource *models.Entity) error
}

// Decide asks the PDP once whether caller may perform action on resource, and records the
// caller when it may (ADR 0004). It returns nil when permitted, ErrForbidden when not, and
// ErrPrincipalNotRecorded when a permitted non-read caller could not be recorded.
func (a *auth) Decide(ctx context.Context, caller *RequestPrincipal, action string, resource *models.Entity) error {
	if resource == nil {
		return fmt.Errorf("decide: resource is required")
	}

	uid := uuid.New()
	parc := ManagementPARC(ctx, caller, action, resource)

	allowed := a.noAuth

	if !allowed && a.pdp != nil {
		resp, err := a.pdp.Authorize(uid.String(), parc)
		if err != nil {
			return err
		}

		allowed = resp != nil && resp.Allowed
	}

	if !allowed {
		a.log.Warn("authorization denied", "uid", uid.String(), "principal", caller.Label(), "action", action, "resource", resource.UID())
		return ErrForbidden
	}

	if a.debug {
		a.log.Debug("authorization granted", "uid", uid.String(), "principal", caller.Label(), "action", action, "resource", resource.UID())
	}

	return a.recordPrincipal(action == models.ActionRead, &caller.Principal)
}

// ManagementPARC builds the PDP request of the management plane (ADR 0006): the caller as
// subject with only its roles, the action as `Action::"<name>"`, the given resource, and a
// context limited to the time and the trace of the request.
func ManagementPARC(ctx context.Context, caller *RequestPrincipal, action string, resource *models.Entity) *models.PARC {
	principal := models.NewEntity(caller.Kind.String(), caller.ID, models.NewAttributeSet())
	if len(caller.Roles) > 0 {
		principal.Attributes().AddAttributeKV(models.AttrRoles, caller.Roles)
	}

	c := models.NewAttributeSet()
	c.AddAttributeKV(models.AttrTime, time.Now().UTC())

	if t := TraceFromContext(ctx); t.Parent != "" {
		c.AddAttributeKV(models.AttrTraceParent, t.Parent)

		if t.State != "" {
			c.AddAttributeKV(models.AttrTraceState, t.State)
		}
	}

	return &models.PARC{
		Principal: principal,
		Action:    models.NewEntity(models.EntityTypeManagementAction, action, models.NewAttributeSet()),
		Resource:  resource,
		Context:   c,
	}
}
