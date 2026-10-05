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

// TestLanguageAndDecisionMatrix pins the role matrix for the policy languages and the
// Authorization Decision Log, decided by the shipped policies: every role reads both, nobody
// changes either, and an API key reads the languages but not the log (SCRUM-146).
func TestLanguageAndDecisionMatrix(t *testing.T) {
	t.Parallel()

	d := managerAuthorizer(t)
	app := &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "key-1")}

	languages := languageCollection()
	decisions := decisionCollection()

	cases := []struct {
		name     string
		caller   *authorization.RequestPrincipal
		action   string
		resource *models.Entity
		want     bool
	}{
		// everyone reads, the ADL included (ADR 0006).
		{"auditor reads languages", withRole("auditor"), models.ActionRead, languages, true},
		{"author reads languages", withRole("author"), models.ActionRead, languages, true},
		{"admin reads languages", withRole("admin"), models.ActionRead, languages, true},
		{"auditor reads decisions", withRole("auditor"), models.ActionRead, decisions, true},
		{"author reads decisions", withRole("author"), models.ActionRead, decisions, true},
		{"admin reads decisions", withRole("admin"), models.ActionRead, decisions, true},

		// neither type has a write operation, and no rule grants one.
		{"admin must NOT update languages", withRole("admin"), models.ActionUpdate, languages, false},
		{"author must NOT delete decisions", withRole("author"), models.ActionDelete, decisions, false},
		{"admin must NOT delete decisions", withRole("admin"), models.ActionDelete, decisions, false},

		// locally seeded users carry roles as a plain string, not a set.
		{"scalar author role reads languages", scalarRole("author"), models.ActionRead, languages, true},
		{"scalar auditor role reads decisions", scalarRole("auditor"), models.ActionRead, decisions, true},

		// no role, an unknown role, the dropped user role, or no identity at all: nothing.
		{"user without roles must NOT read languages", withRole(""), models.ActionRead, languages, false},
		{"unknown role must NOT read decisions", withRole("visitor"), models.ActionRead, decisions, false},
		{"user role must NOT read languages", withRole("user"), models.ActionRead, languages, false},
		{"user role must NOT read decisions", withRole("user"), models.ActionRead, decisions, false},
		{"system must NOT read decisions", authorization.SystemRequestPrincipal(), models.ActionRead, decisions, false},

		// an API key fetches bundles: it reads the languages, never the log.
		{"app reads languages", app, models.ActionRead, languages, true},
		{"app must NOT read decisions", app, models.ActionRead, decisions, false},
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
