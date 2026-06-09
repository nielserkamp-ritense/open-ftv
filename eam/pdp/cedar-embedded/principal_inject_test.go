package cedar_embedded

import (
	"log/slog"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestAuthorizeUsesRequestSuppliedRoles(t *testing.T) {
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

	c := NewController(pdp.WithLogger(logger)).(*controller)

	var p cedar.Policy
	require.NoError(t, p.UnmarshalCedar([]byte(
		`permit (principal is user, action, resource is service) when { principal has roles && principal.roles.contains("admin") };`,
	)))
	c.pdp.Add(cedar.PolicyID("admin-roles"), &p)

	principal := models.NewEntity("user", "alice", models.NewAttributeSet())
	principal.Attributes().AddAttributeKV("roles", []string{"admin"})

	parc := &models.PARC{
		Principal: principal,
		Action:    models.NewEntity(models.EntityTypeName, "can_read", models.NewAttributeSet()),
		Resource:  models.NewEntity(models.EntityTypeService, "https://mgr.test/v1/policies", models.NewAttributeSet()),
		Context:   models.NewAttributeSet(),
	}

	resp, err := c.Authorize("test-uid", parc)
	require.NoError(t, err)
	assert.True(t, resp.Allowed, "principal with request-supplied roles=[admin] must be permitted")
}
