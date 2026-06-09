package cedar_embedded

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func managerPDP(t *testing.T) *controller {
	t.Helper()
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	ep := pep.New(nil, logger)
	ip := pip.New(nil, logger)
	ap := pap.New(nil, logger, pap.WithLanguage("cedar"),
		pap.WithFileStore("../../../testdata/apps/manager/policies/cedar", true))
	c := NewController(pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
	return c.(*controller)
}

func parcFor(role, action, path string) *models.PARC {
	principal := models.NewEntity("user", role+"-user", models.NewAttributeSet())
	principal.Attributes().AddAttributeKV("roles", []string{role})
	ctx := models.NewAttributeSet()
	ctx.AddAttributeKV(models.AttrHTTP, map[string]any{models.AttrPath: path})
	return &models.PARC{
		Principal: principal,
		Action:    models.NewEntity(models.EntityTypeName, action, models.NewAttributeSet()),
		Resource:  models.NewEntity(models.EntityTypeService, "https://mgr.test"+path, models.NewAttributeSet()),
		Context:   ctx,
	}
}

func TestRolePolicyMatrix(t *testing.T) {
	c := managerPDP(t)

	cases := []struct {
		role, action, path string
		want               bool
	}{
		{"auditor", "can_read", "/v1/adl/entries", true},
		{"auditor", "can_update", "/v1/policy", false},
		{"auditor", "can_create", "/v1/policy", false},
		{"author", "can_read", "/v1/policies", true},
		{"author", "can_update", "/v1/policy", true},
		{"author", "can_create", "/v1/attribute", true},
		{"author", "can_delete", "/v1/tag", true},
		{"admin", "can_read", "/v1/policies", true},
		{"admin", "can_update", "/v1/policy", true},
		{"author", "can_update", "/v1/deployment", false},
		{"auditor", "can_update", "/v1/deployment", false},
		{"admin", "can_update", "/v1/deployment", true},
		{"author", "can_update", "/v1/policy", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.role+"-"+tc.action+"-"+tc.path, func(t *testing.T) {
			resp, err := c.Authorize("uid", parcFor(tc.role, tc.action, tc.path))
			require.NoError(t, err)
			assert.Equal(t, tc.want, resp.Allowed)
		})
	}
}

func TestDraftGuard(t *testing.T) {
	c := managerPDP(t) // loads testdata/apps/manager/policies/cedar (incl. policy_draft_guard.cedar)

	mk := func(role, status string) *models.PARC {
		pr := models.NewEntity("user", role+"-u", models.NewAttributeSet())
		pr.Attributes().AddAttributeKV("roles", []string{role})
		res := models.NewEntity(models.EntityTypeService, "https://mgr.test/v1/policy/x", models.NewAttributeSet())
		res.Attributes().AddAttributeKV("status", status)
		ctx := models.NewAttributeSet()
		ctx.AddAttributeKV(models.AttrHTTP, map[string]any{models.AttrPath: "/v1/policy/x"})
		return &models.PARC{
			Principal: pr,
			Action:    models.NewEntity(models.EntityTypeName, "can_update", models.NewAttributeSet()),
			Resource:  res,
			Context:   ctx,
		}
	}

	cases := []struct {
		role, status string
		want         bool
	}{
		{"author", "draft", true},      // author may update a draft
		{"author", "published", false}, // author may NOT update a published policy
		{"admin", "published", true},   // admin may update any status
		{"admin", "draft", true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.role+"-"+tc.status, func(t *testing.T) {
			resp, err := c.Authorize("u", mk(tc.role, tc.status))
			require.NoError(t, err)
			assert.Equal(t, tc.want, resp.Allowed, "%s updating %s policy", tc.role, tc.status)
		})
	}
}
