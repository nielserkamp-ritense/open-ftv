package management

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// TestAttributeAndEntityMatrix pins the SCRUM-16 role matrix for PIP data (attributes and
// entities), decided by the shipped policies: the author creates, updates, deletes, accepts,
// reverts and restores; every role reads; the admin and the auditor change nothing; an
// accepted object is not updated in place by anyone, and an edit of deployed data is allowed
// because it is the only way its next concept starts.
func TestAttributeAndEntityMatrix(t *testing.T) {
	t.Parallel()

	d := managerAuthorizer(t)
	app := &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "key-1")}

	attrs := attributeCollection()
	concept := pipResource(models.EntityTypeAttribute, "max-age", "concept", []string{"laadpalen"})
	accepted := pipResource(models.EntityTypeAttribute, "max-age", "accepted", []string{"laadpalen"})
	missing := attributeResource("nope", nil)

	entities := entityCollection()
	eConcept := pipResource(models.EntityTypeEntity, "user::alice", "concept", nil)
	eAccepted := pipResource(models.EntityTypeEntity, "user::alice", "accepted", nil)
	deployed := pipResource(models.EntityTypeAttribute, "max-age", "deployed", []string{"laadpalen"})
	eDeployed := pipResource(models.EntityTypeEntity, "user::alice", "deployed", nil)

	cases := []struct {
		name     string
		caller   *authorization.RequestPrincipal
		action   string
		resource *models.Entity
		want     bool
	}{
		// everyone reads, single objects and the collection alike.
		{"auditor reads attributes", withRole("auditor"), models.ActionRead, attrs, true},
		{"auditor reads attribute", withRole("auditor"), models.ActionRead, concept, true},
		{"author reads attribute", withRole("author"), models.ActionRead, concept, true},
		{"admin reads attribute", withRole("admin"), models.ActionRead, concept, true},
		{"admin reads missing attribute", withRole("admin"), models.ActionRead, missing, true},
		{"auditor reads entities", withRole("auditor"), models.ActionRead, entities, true},
		{"author reads entity", withRole("author"), models.ActionRead, eConcept, true},
		{"admin reads entity", withRole("admin"), models.ActionRead, eConcept, true},

		// the functioneel beheerder (author) administers PIP data.
		{"author creates attribute", withRole("author"), models.ActionCreate, concept, true},
		{"author updates concept attribute", withRole("author"), models.ActionUpdate, concept, true},
		{"author deletes attribute", withRole("author"), models.ActionDelete, concept, true},
		{"author accepts attribute", withRole("author"), models.ActionAccept, concept, true},
		{"author reverts attribute", withRole("author"), models.ActionRevert, accepted, true},
		{"author restores attribute", withRole("author"), models.ActionRestore, accepted, true},
		{"author deletes accepted attribute", withRole("author"), models.ActionDelete, accepted, true},
		{"author creates entity", withRole("author"), models.ActionCreate, eConcept, true},
		{"author updates concept entity", withRole("author"), models.ActionUpdate, eConcept, true},
		{"author deletes entity", withRole("author"), models.ActionDelete, eConcept, true},
		{"author accepts entity", withRole("author"), models.ActionAccept, eConcept, true},
		{"author reverts entity", withRole("author"), models.ActionRevert, eAccepted, true},
		{"author restores entity", withRole("author"), models.ActionRestore, eAccepted, true},

		// an accepted object is not edited in place: revert it first (pip_accepted_guard).
		{"author must NOT update accepted attribute", withRole("author"), models.ActionUpdate, accepted, false},
		{"author must NOT update accepted entity", withRole("author"), models.ActionUpdate, eAccepted, false},

		// deployed data has no way back to concept, so its edit is not caught by the guard; the
		// roles that may not update still may not.
		{"author updates deployed attribute", withRole("author"), models.ActionUpdate, deployed, true},
		{"author updates deployed entity", withRole("author"), models.ActionUpdate, eDeployed, true},
		{"admin must NOT update deployed attribute", withRole("admin"), models.ActionUpdate, deployed, false},
		{"auditor must NOT update deployed entity", withRole("auditor"), models.ActionUpdate, eDeployed, false},

		// the systeembeheerder (admin) administers beslispunten and settings, not PIP data:
		// this is what admin.cedar allowed.
		{"admin must NOT create attribute", withRole("admin"), models.ActionCreate, concept, false},
		{"admin must NOT update attribute", withRole("admin"), models.ActionUpdate, concept, false},
		{"admin must NOT delete attribute", withRole("admin"), models.ActionDelete, concept, false},
		{"admin must NOT accept attribute", withRole("admin"), models.ActionAccept, concept, false},
		{"admin must NOT restore attribute", withRole("admin"), models.ActionRestore, accepted, false},
		{"admin must NOT create entity", withRole("admin"), models.ActionCreate, eConcept, false},
		{"admin must NOT accept entity", withRole("admin"), models.ActionAccept, eConcept, false},

		// the auditor never changes anything.
		{"auditor must NOT create attribute", withRole("auditor"), models.ActionCreate, concept, false},
		{"auditor must NOT update attribute", withRole("auditor"), models.ActionUpdate, concept, false},
		{"auditor must NOT delete attribute", withRole("auditor"), models.ActionDelete, concept, false},
		{"auditor must NOT accept attribute", withRole("auditor"), models.ActionAccept, concept, false},
		{"auditor must NOT revert attribute", withRole("auditor"), models.ActionRevert, accepted, false},
		{"auditor must NOT restore attribute", withRole("auditor"), models.ActionRestore, accepted, false},
		{"auditor must NOT update entity", withRole("auditor"), models.ActionUpdate, eConcept, false},
		{"auditor must NOT accept entity", withRole("auditor"), models.ActionAccept, eConcept, false},

		// locally seeded users carry roles as a plain string, not a set.
		{"scalar author role updates attribute", scalarRole("author"), models.ActionUpdate, concept, true},
		{"scalar auditor role reads entity", scalarRole("auditor"), models.ActionRead, eConcept, true},

		// no role, an unknown role, or no identity at all: nothing.
		{"user without roles must NOT read attribute", withRole(""), models.ActionRead, concept, false},
		{"unknown role must NOT read entity", withRole("visitor"), models.ActionRead, eConcept, false},
		{"user role must NOT read attribute", withRole("user"), models.ActionRead, concept, false},
		{"user role must NOT read entities", withRole("user"), models.ActionRead, entities, false},
		{"system must NOT read attribute", authorization.SystemRequestPrincipal(), models.ActionRead, concept, false},

		// an API key reads only.
		{"app reads attribute", app, models.ActionRead, concept, true},
		{"app reads entities", app, models.ActionRead, entities, true},
		{"app must NOT create attribute", app, models.ActionCreate, concept, false},
		{"app must NOT update entity", app, models.ActionUpdate, eConcept, false},
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
