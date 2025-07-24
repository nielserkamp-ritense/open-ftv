package fiber

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewTagsHandler(t *testing.T) {
	t.Parallel()

	t.Run("new tags handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		tags := []*policies.Tag{
			{Id: "ux", Name: "User interface"},
			{Id: "pap", Name: "PAP processen"},
			{Id: "pip", Name: "PIP processen"},
			{Id: "bronnen", Name: "Externe bronnen"},
		}

		th := NewTagsHandler(logger, tags, nil)
		require.NotNil(t, th)
	})
}

func TestTagsHandler_GetPolicies(t *testing.T) {
	t.Parallel()

	t.Run("get tags", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		tags := []*policies.Tag{
			{Id: "ux", Name: "User interface"},
			{Id: "pap", Name: "PAP processen"},
			{Id: "pip", Name: "PIP processen"},
			{Id: "bronnen", Name: "Externe bronnen"},
		}

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ph := NewTagsHandler(logger, tags, auth)
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/tags", ph.GetTags)

		req := httptest.NewRequest(fiber.MethodGet, "/v1/tags", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, PoliciesVersion, resp.Header.Get(HeaderVersion))
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list []*policies.Tag
		err := json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.Equal(t, 4, len(list))
	})
}
