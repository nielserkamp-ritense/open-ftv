package fiber

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// TestIdentify_handsTheCallerToTheHandler proves the middleware validates the token once and
// leaves the request-scoped Principal, with roles and trace, in the handler's context.
func TestIdentify_handsTheCallerToTheHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	ep := pep.New(ctx, logger, pep.WithJWT(pep.JWTConfig{
		Keyfunc:  func(*jwt.Token) (any, error) { return &key.PublicKey, nil },
		Issuer:   "https://idp.test",
		Audience: "openftv",
	}))
	a := authorization.New(authorization.WithContext(ctx), authorization.WithLogger(logger), authorization.WithPEP(ep), authorization.NoAuth())

	var (
		caller *authorization.RequestPrincipal
		found  bool
		trace  authorization.Trace
	)

	app := fiber.New()
	app.Get("/v1/policies", Identify(a, logger), func(c *fiber.Ctx) error {
		caller, found = authorization.RequestPrincipalFromContext(c.UserContext())
		trace = authorization.TraceFromContext(c.UserContext())

		return c.SendStatus(fiber.StatusNoContent)
	})

	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":                "https://idp.test",
		"aud":                "openftv",
		"sub":                "alice",
		"preferred_username": "Alice",
		"roles":              []string{"author", "auditor"},
		"exp":                time.Now().Add(time.Hour).Unix(),
	})
	bearer, err := tok.SignedString(key)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/v1/policies", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+bearer)
	req.Header.Set("traceparent", "00-abc-def-01")

	res, err := app.Test(req)
	require.NoError(t, err)

	defer res.Body.Close()

	require.Equal(t, fiber.StatusNoContent, res.StatusCode)

	require.True(t, found)
	assert.Equal(t, identity.KindUser, caller.Kind)
	assert.Equal(t, "alice", caller.ID)
	assert.Equal(t, "Alice", caller.Name)
	assert.Equal(t, []string{"author", "auditor"}, caller.Roles)
	assert.Equal(t, "00-abc-def-01", trace.Parent)
}

// TestIdentify_noCredentialsIsTheSystem: a request without any identity passes as the system
// principal, so the PDP (not the middleware) decides what an anonymous caller may do.
func TestIdentify_noCredentialsIsTheSystem(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))
	a := authorization.New(authorization.WithLogger(logger), authorization.NoAuth())

	var caller *authorization.RequestPrincipal

	app := fiber.New()
	app.Get("/v1/policies", Identify(a, logger), func(c *fiber.Ctx) error {
		caller, _ = authorization.RequestPrincipalFromContext(c.UserContext())

		return c.SendStatus(fiber.StatusNoContent)
	})

	res, err := app.Test(httptest.NewRequest(http.MethodGet, "/v1/policies", http.NoBody))
	require.NoError(t, err)

	defer res.Body.Close()

	require.Equal(t, fiber.StatusNoContent, res.StatusCode)
	assert.Equal(t, identity.KindSystem, caller.Kind)
}
