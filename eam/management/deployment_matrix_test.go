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

func storedDeployment(status string) *models.Entity {
	attrs := models.NewAttributeSet()
	attrs.AddAttributeKV(models.AttrStatus, status)

	return models.NewEntity(models.EntityTypeDeployment, "7", attrs)
}

// TestDeploymentMatrix pins the SCRUM-16 role matrix for deployments, decided by the shipped
// policies: everyone reads, the author publishes, admin and auditor do not.
func TestDeploymentMatrix(t *testing.T) {
	t.Parallel()

	d := managerAuthorizer(t)
	completed := storedDeployment("completed")
	collection := models.NewEntity(models.EntityTypeDeployment, models.ResourceCollection, models.NewAttributeSet())
	app := &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "key-1")}

	cases := []struct {
		name     string
		caller   *authorization.RequestPrincipal
		action   string
		resource *models.Entity
		want     bool
	}{
		// everyone reads, single deployments and the collection alike.
		{"auditor reads collection", withRole("auditor"), models.ActionRead, collection, true},
		{"auditor reads deployment", withRole("auditor"), models.ActionRead, completed, true},
		{"admin reads deployment", withRole("admin"), models.ActionRead, completed, true},
		{"author reads deployment", withRole("author"), models.ActionRead, completed, true},

		// the functioneel beheerder (author) publishes.
		{"author deploys", withRole("author"), models.ActionDeploy, collection, true},

		// the systeembeheerder (admin) no longer publishes; the auditor never did.
		{"admin must NOT deploy", withRole("admin"), models.ActionDeploy, collection, false},
		{"auditor must NOT deploy", withRole("auditor"), models.ActionDeploy, collection, false},

		// there is no editing of a deployment for anyone.
		{"author must NOT update", withRole("author"), models.ActionUpdate, completed, false},
		{"author must NOT delete", withRole("author"), models.ActionDelete, completed, false},

		// locally seeded users carry roles as a plain string, not a set.
		{"scalar author role deploys", scalarRole("author"), models.ActionDeploy, collection, true},
		{"scalar auditor role reads", scalarRole("auditor"), models.ActionRead, completed, true},

		// no role, an unknown role, or no identity at all: nothing.
		{"user without roles must NOT read", withRole(""), models.ActionRead, completed, false},
		{"unknown role must NOT read", withRole("visitor"), models.ActionRead, completed, false},
		{"system must NOT read", authorization.SystemRequestPrincipal(), models.ActionRead, completed, false},

		// a PDP fetches its bundle with an API key: read only.
		{"app reads collection", app, models.ActionRead, collection, true},
		{"app reads deployment", app, models.ActionRead, completed, true},
		{"app must NOT deploy", app, models.ActionDeploy, collection, false},
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
