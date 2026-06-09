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

func TestAuthorizeUsesResourceAttributes(t *testing.T) {
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	c := NewController(pdp.WithLogger(logger)).(*controller)

	var p cedar.Policy
	require.NoError(t, p.UnmarshalCedar([]byte(
		`permit (principal, action, resource is service) when { resource has status && resource.status == "draft" };`)))
	c.pdp.Add(cedar.PolicyID("draft-only"), &p)

	mk := func(status string) *models.PARC {
		res := models.NewEntity(models.EntityTypeService, "p1", models.NewAttributeSet())
		res.Attributes().AddAttributeKV("status", status)
		return &models.PARC{
			Principal: models.NewEntity("user", "alice", models.NewAttributeSet()),
			Action:    models.NewEntity(models.EntityTypeName, "can_update", models.NewAttributeSet()),
			Resource:  res,
			Context:   models.NewAttributeSet(),
		}
	}
	d, err := c.Authorize("u", mk("draft"))
	require.NoError(t, err)
	assert.True(t, d.Allowed, "draft must permit")
	d2, err := c.Authorize("u", mk("published"))
	require.NoError(t, err)
	assert.False(t, d2.Allowed, "published must deny")
}
