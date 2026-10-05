package management

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

const (
	papSeedDir = "../../testdata/apps/pap/policies"
	pipSeedDir = "../../testdata/apps/pip/policies"
)

// oldModel matches a rule written in the HTTP-shaped model ADR 0006 replaced: the `service`
// resource type or a verb-derived `can_*` action. Prose in a comment does not match.
var oldModel = regexp.MustCompile(`resource\s+is\s+service\b|name::"can_`)

// TestSeedPolicies_noOldModelRules guards the end of the rollout (SCRUM-146): no shipped
// policy of the manager, the PAP or the PIP is written in the old vocabulary any more.
func TestSeedPolicies_noOldModelRules(t *testing.T) {
	t.Parallel()

	for _, dir := range []string{"../../testdata/apps/manager/policies/cedar", papSeedDir, pipSeedDir} {
		files, err := filepath.Glob(filepath.Join(dir, "*.cedar"))
		require.NoError(t, err)
		require.NotEmpty(t, files, dir)

		for _, f := range files {
			b, err := os.ReadFile(f)
			require.NoError(t, err)
			assert.Falsef(t, oldModel.Match(b), "%s still matches on the old model", f)
		}
	}
}

// TestStandaloneSeedMatrix pins the shipped policies of the standalone PAP and PIP. They keep
// their own role semantics (ADR 0006): admin does everything, user and an API key read.
func TestStandaloneSeedMatrix(t *testing.T) {
	t.Parallel()

	collection := func(typ string) *models.Entity {
		return models.NewEntity(typ, models.ResourceCollection, models.NewAttributeSet())
	}

	suites := []struct {
		name  string
		dir   string
		types []string
		other string // a type this app does not administer: no rule covers it.
	}{
		{"pap", papSeedDir, []string{models.EntityTypePolicy, models.EntityTypeTag, models.EntityTypeDeployment, models.EntityTypeLanguage}, models.EntityTypeAttribute},
		{"pip", pipSeedDir, []string{models.EntityTypeAttribute, models.EntityTypeEntity}, models.EntityTypePolicy},
	}

	app := &authorization.RequestPrincipal{Principal: identity.NewPrincipal(identity.KindApp, "key-1")}

	for _, s := range suites {
		t.Run(s.name, func(t *testing.T) {
			t.Parallel()

			d := seedAuthorizer(t, s.dir)

			check := func(name string, caller *authorization.RequestPrincipal, action string, resource *models.Entity, want bool) {
				t.Helper()

				err := d.Authorize(context.Background(), caller, action, resource)
				if want {
					require.NoError(t, err, name)
				} else {
					require.ErrorIs(t, err, authorization.ErrForbidden, name)
				}
			}

			for _, typ := range s.types {
				r := collection(typ)

				check("admin reads "+typ, withRole("admin"), models.ActionRead, r, true)
				check("admin creates "+typ, withRole("admin"), models.ActionCreate, r, true)
				check("admin deletes "+typ, withRole("admin"), models.ActionDelete, r, true)

				check("user reads "+typ, withRole("user"), models.ActionRead, r, true)
				check("user must NOT create "+typ, withRole("user"), models.ActionCreate, r, false)
				check("user must NOT delete "+typ, withRole("user"), models.ActionDelete, r, false)

				check("app reads "+typ, app, models.ActionRead, r, true)
				check("app must NOT create "+typ, app, models.ActionCreate, r, false)

				check("unknown role must NOT read "+typ, withRole("visitor"), models.ActionRead, r, false)
				check("user without roles must NOT read "+typ, withRole(""), models.ActionRead, r, false)
			}

			// nothing is left that grants across types, as the old admin rule did.
			check("admin must NOT read "+s.other, withRole("admin"), models.ActionRead, collection(s.other), false)
			check("admin must NOT read decision", withRole("admin"), models.ActionRead, decisionCollection(), false)
		})
	}
}
