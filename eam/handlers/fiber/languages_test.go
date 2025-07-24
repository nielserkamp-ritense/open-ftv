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

func TestNewLanguagesHandler(t *testing.T) {
	t.Parallel()

	t.Run("new languages handler", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		lh := NewLanguagesHandler(logger, nil)
		require.NotNil(t, lh)
	})
}

func TestLanguagesHandler_GetPolicies(t *testing.T) {
	t.Parallel()

	t.Run("get languages", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ph := NewLanguagesHandler(logger, auth)
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/languages", ph.GetLanguages)

		req := httptest.NewRequest(fiber.MethodGet, "/v1/languages", nil)
		resp, err2 := srv.Test(req, 100)

		require.NoError(t, err2)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, PoliciesVersion, resp.Header.Get(HeaderVersion))
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err3 := io.ReadAll(resp.Body)
		require.NoError(t, err3)
		require.NotNil(t, b)

		var list []*policies.Language
		err := json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 4)
	})
}
