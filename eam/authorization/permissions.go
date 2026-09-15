package authorization

import (
	"slices"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Permissions is the management plane's role table: which actions a user with a role, and an
// API key, may perform. It is the decision behind Authorize until the embedded PDP decides on
// the management-plane model (ADR 0006); the table encodes the same role matrix the shipped
// example policies will.
type Permissions struct {
	// Roles maps a role to the actions a user carrying it may perform.
	Roles map[string][]string
	// App lists the actions an API key may perform.
	App []string
}

// Allows reports whether caller may perform action. A user is allowed when any of its roles
// allows the action; an app when the action is in App; anything else, including the system
// principal and a user without roles, is never allowed.
func (p *Permissions) Allows(caller *RequestPrincipal, action string) bool {
	if p == nil || caller == nil {
		return false
	}

	switch {
	case caller.IsAuthenticatedUser():
		for _, role := range caller.Roles {
			if slices.Contains(p.Roles[role], action) {
				return true
			}
		}

		return false

	case caller.Kind == identity.KindApp && caller.ID != "":
		return slices.Contains(p.App, action)

	default:
		return false
	}
}

// Role names the shipped tables know (ADR 0002 claim values; the PAP's local `user`).
const (
	RoleAdmin   = "admin"
	RoleAuthor  = "author"
	RoleAuditor = "auditor"
	RoleUser    = "user"
)

// policyActions are all actions the policy service decides.
var policyActions = []string{
	models.ActionRead, models.ActionCreate, models.ActionUpdate, models.ActionDelete,
	models.ActionAccept, models.ActionRevert, models.ActionDeploy, models.ActionRestore,
}

// ManagerPermissions is the SCRUM-16 role matrix for the manager's policies: the functioneel
// beheerder (author) edits, accepts, reverts, deploys and restores; the systeembeheerder
// (admin) and the auditor read; an API key reads.
func ManagerPermissions() *Permissions {
	return &Permissions{
		Roles: map[string][]string{
			RoleAuthor:  policyActions,
			RoleAdmin:   {models.ActionRead},
			RoleAuditor: {models.ActionRead},
		},
		App: []string{models.ActionRead},
	}
}

// PAPPermissions is the standalone PAP's role table (ADR 0006 names it as the exception to
// the matrix): its locally seeded users carry `admin` and `user`, so admin does everything
// and user, like an API key, reads.
func PAPPermissions() *Permissions {
	return &Permissions{
		Roles: map[string][]string{
			RoleAdmin: policyActions,
			RoleUser:  {models.ActionRead},
		},
		App: []string{models.ActionRead},
	}
}
