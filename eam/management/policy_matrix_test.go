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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	cedar "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/cedar-embedded"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// managerDecider is the real chain: the authorizer's Decide over the embedded Cedar PDP
// loaded with the manager's shipped policies.
func managerDecider(t *testing.T) authorization.Decider {
	t.Helper()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	ctx := context.Background()

	store := pap.New(ctx, logger, pap.WithLanguage("cedar"),
		pap.WithFileStore("../../testdata/apps/manager/policies/cedar", true))

	// Users the PIP holds reach the PDP with the PIP's attributes, where `roles` may be one
	// string; the seed rules must cope with both shapes.
	ip := pip.New(ctx, logger)

	for _, role := range []string{"author", "auditor"} {
		attrs := models.NewAttributeSet()
		attrs.AddAttributeKV(models.AttrRoles, role)
		_, err := ip.AddEntity(models.NewEntity("user", role+"-scalar", attrs))
		require.NoError(t, err)
	}

	controller := cedar.NewController(
		pdp.WithContext(ctx), pdp.WithLogger(logger),
		pdp.WithPEP(pep.New(ctx, logger)), pdp.WithPIP(ip), pdp.WithPAP(store))

	return authorization.New(authorization.WithContext(ctx), authorization.WithLogger(logger), authorization.WithPDP(controller))
}

func withRole(role string) *authorization.RequestPrincipal {
	p := &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindUser, role+"-user")}
	if role != "" {
		p.Roles = []string{role}
	}

	return p
}

// scalarRole builds a caller whose roles reach the PDP as one string rather than a set, as
// happens for users the PIP holds with `roles: "author"`.
func scalarRole(role string) *authorization.RequestPrincipal {
	return &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindUser, role+"-scalar"), Roles: []string{role}}
}

func storedPolicy(status string) *models.Entity {
	attrs := models.NewAttributeSet()
	attrs.AddAttributeKV(models.AttrStatus, status)
	attrs.AddAttributeKV(models.AttrTags, []string{"laadpalen"})
	attrs.AddAttributeKV(models.AttrLanguage, "cedar")

	return models.NewEntity(models.EntityTypePolicy, "p1", attrs)
}

// TestPolicyMatrix pins the SCRUM-16 role matrix for policies, decided by the shipped
// policies: author edits, accepts, deploys and restores; admin and auditor only read.
func TestPolicyMatrix(t *testing.T) {
	t.Parallel()

	d := managerDecider(t)
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

		// an accepted policy is not edited in place.
		{"author must NOT update accepted", withRole("author"), models.ActionUpdate, accepted, false},

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

		// locally seeded users carry roles as a plain string, not a set.
		{"scalar author role edits", scalarRole("author"), models.ActionUpdate, concept, true},
		{"scalar auditor role reads", scalarRole("auditor"), models.ActionRead, concept, true},

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

			err := d.Decide(context.Background(), tc.caller, tc.action, tc.resource)

			if tc.want {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, authorization.ErrForbidden)
			}
		})
	}
}
