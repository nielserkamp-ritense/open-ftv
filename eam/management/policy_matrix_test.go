package management

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// managerAuthorizer is the authorizer as the manager wires it: Authorize over the manager's
// role table.
func managerAuthorizer() authorization.Authorizer {
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

	return authorization.New(authorization.WithLogger(logger), authorization.WithPermissions(authorization.ManagerPermissions()))
}

func withRole(role string) *authorization.RequestPrincipal {
	p := &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindUser, role+"-user")}
	if role != "" {
		p.Roles = []string{role}
	}

	return p
}

func storedPolicy(status string) *models.Entity {
	attrs := models.NewAttributeSet()
	attrs.AddAttributeKV(models.AttrStatus, status)
	attrs.AddAttributeKV(models.AttrTags, []string{"laadpalen"})
	attrs.AddAttributeKV(models.AttrLanguage, "cedar")

	return models.NewEntity(models.EntityTypePolicy, "p1", attrs)
}

// TestPolicyMatrix pins the SCRUM-16 role matrix for policies as the manager decides it:
// author edits, accepts, reverts, deploys and restores; admin and auditor only read.
func TestPolicyMatrix(t *testing.T) {
	t.Parallel()

	d := managerAuthorizer()
	concept, accepted := storedPolicy("concept"), storedPolicy("accepted")
	collection := models.NewEntity(models.EntityTypePolicy, models.ResourceCollection, models.NewAttributeSet())

	cases := []struct {
		name     string
		caller   *authorization.RequestPrincipal
		action   string
		resource *models.Entity
		want     bool
	}{
		// everyone reads, single objects and the collection alike.
		{"auditor reads collection", withRole("auditor"), models.ActionRead, collection, true},
		{"auditor reads policy", withRole("auditor"), models.ActionRead, concept, true},
		{"admin reads policy", withRole("admin"), models.ActionRead, concept, true},
		{"author reads policy", withRole("author"), models.ActionRead, concept, true},

		// the functioneel beheerder (author) administers rules.
		{"author creates", withRole("author"), models.ActionCreate, concept, true},
		{"author updates concept", withRole("author"), models.ActionUpdate, concept, true},
		{"author deletes", withRole("author"), models.ActionDelete, concept, true},
		{"author accepts", withRole("author"), models.ActionAccept, concept, true},
		{"author deploys", withRole("author"), models.ActionDeploy, accepted, true},
		{"author restores accepted", withRole("author"), models.ActionRestore, accepted, true},
		{"author reverts accepted", withRole("author"), models.ActionRevert, accepted, true},

		// the auditor never changes anything: this is the PATCH hole, closed.
		{"auditor must NOT update", withRole("auditor"), models.ActionUpdate, concept, false},
		{"auditor must NOT accept", withRole("auditor"), models.ActionAccept, concept, false},
		{"auditor must NOT deploy", withRole("auditor"), models.ActionDeploy, accepted, false},
		{"auditor must NOT revert", withRole("auditor"), models.ActionRevert, accepted, false},
		{"auditor must NOT create", withRole("auditor"), models.ActionCreate, concept, false},
		{"auditor must NOT delete", withRole("auditor"), models.ActionDelete, concept, false},

		// the systeembeheerder (admin) is not a superuser: no rule editing, no deployment.
		{"admin must NOT create", withRole("admin"), models.ActionCreate, concept, false},
		{"admin must NOT update", withRole("admin"), models.ActionUpdate, concept, false},
		{"admin must NOT accept", withRole("admin"), models.ActionAccept, concept, false},
		{"admin must NOT deploy", withRole("admin"), models.ActionDeploy, accepted, false},
		{"admin must NOT revert", withRole("admin"), models.ActionRevert, accepted, false},
		{"admin must NOT delete", withRole("admin"), models.ActionDelete, concept, false},

		// no role, an unknown role, or no identity at all: nothing.
		{"user without roles must NOT read", withRole(""), models.ActionRead, concept, false},
		{"unknown role must NOT read", withRole("visitor"), models.ActionRead, concept, false},
		{"system must NOT read", authorization.SystemRequestPrincipal(), models.ActionRead, concept, false},

		// an API key reads only.
		{"app reads", &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "key-1")}, models.ActionRead, concept, true},
		{"app must NOT update", &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "key-1")}, models.ActionUpdate, concept, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := d.Authorize(context.Background(), tc.caller, tc.action, tc.resource)

			if tc.want {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, authorization.ErrForbidden)
			}
		})
	}
}
