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

// TestTagAndSettingMatrix pins the SCRUM-16 role matrix for beslispunten (tags) and settings,
// decided by the shipped policies: admin creates, updates and deletes; every role reads; the
// author and the auditor change nothing.
func TestTagAndSettingMatrix(t *testing.T) {
	t.Parallel()

	d := managerAuthorizer(t)
	app := &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "key-1")}

	tag := tagResource("laadpalen")
	tags := tagCollection()
	setting := settingResource(SettingUI)

	cases := []struct {
		name     string
		caller   *authorization.RequestPrincipal
		action   string
		resource *models.Entity
		want     bool
	}{
		// everyone reads, single objects and the collection alike.
		{"auditor reads tags", withRole("auditor"), models.ActionRead, tags, true},
		{"auditor reads tag", withRole("auditor"), models.ActionRead, tag, true},
		{"author reads tag", withRole("author"), models.ActionRead, tag, true},
		{"admin reads tag", withRole("admin"), models.ActionRead, tag, true},
		{"auditor reads settings", withRole("auditor"), models.ActionRead, setting, true},
		{"author reads settings", withRole("author"), models.ActionRead, setting, true},
		{"admin reads settings", withRole("admin"), models.ActionRead, setting, true},

		// the systeembeheerder (admin) manages beslispunten and settings.
		{"admin creates tag", withRole("admin"), models.ActionCreate, tag, true},
		{"admin updates tag", withRole("admin"), models.ActionUpdate, tag, true},
		{"admin deletes tag", withRole("admin"), models.ActionDelete, tag, true},
		{"admin updates settings", withRole("admin"), models.ActionUpdate, setting, true},

		// the author administers rules, not beslispunten: the old model let an author write both.
		{"author must NOT create tag", withRole("author"), models.ActionCreate, tag, false},
		{"author must NOT update tag", withRole("author"), models.ActionUpdate, tag, false},
		{"author must NOT delete tag", withRole("author"), models.ActionDelete, tag, false},
		{"author must NOT update settings", withRole("author"), models.ActionUpdate, setting, false},

		// the auditor never changes anything.
		{"auditor must NOT create tag", withRole("auditor"), models.ActionCreate, tag, false},
		{"auditor must NOT update tag", withRole("auditor"), models.ActionUpdate, tag, false},
		{"auditor must NOT delete tag", withRole("auditor"), models.ActionDelete, tag, false},
		{"auditor must NOT update settings", withRole("auditor"), models.ActionUpdate, setting, false},

		// locally seeded users carry roles as a plain string, not a set.
		{"scalar author role reads tag", scalarRole("author"), models.ActionRead, tag, true},
		{"scalar auditor role reads settings", scalarRole("auditor"), models.ActionRead, setting, true},

		// no role, an unknown role, or no identity at all: nothing.
		{"user without roles must NOT read tag", withRole(""), models.ActionRead, tag, false},
		{"unknown role must NOT read settings", withRole("visitor"), models.ActionRead, setting, false},
		{"system must NOT read tag", authorization.SystemRequestPrincipal(), models.ActionRead, tag, false},

		// an API key reads only.
		{"app reads tag", app, models.ActionRead, tag, true},
		{"app reads settings", app, models.ActionRead, setting, true},
		{"app must NOT create tag", app, models.ActionCreate, tag, false},
		{"app must NOT update settings", app, models.ActionUpdate, setting, false},
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
