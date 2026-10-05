package fiber

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authlog"
)

// stubSearcher is a test double for search.Searcher.
type stubSearcher struct{ got *search.Criteria }

func (s *stubSearcher) Search(_ context.Context, c *search.Criteria) (authlog.AuthlogEntries, error) {
	s.got = c
	return authlog.AuthlogEntries{}, nil
}

func adlStatus(t *testing.T, searcher search.Searcher, authorizer authorization.RequestAuthorizer, query string) int {
	t.Helper()

	logger := settingsLogger()
	srv := fiber.New(fiber.Config{ErrorHandler: server.ErrorHandler(logger)})
	srv.Get("/v1/adl/entries", NewADLHandler(logger, searcher, authorizer).Search)

	resp, err := srv.Test(httptest.NewRequest(fiber.MethodGet, "/v1/adl/entries"+query, http.NoBody), 60000)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp.StatusCode
}

func TestADLHandler_Search(t *testing.T) {
	t.Parallel()

	searcher := &stubSearcher{}

	assert.Equal(t, fiber.StatusOK, adlStatus(t, searcher, settingsAuth(t), "?subjectId=alice&limit=5"))
	require.NotNil(t, searcher.got)
	assert.Equal(t, "alice", searcher.got.SubjectId)
	assert.Equal(t, 5, searcher.got.Limit)
}

func TestADLHandler_Search_Forbidden(t *testing.T) {
	t.Parallel()

	searcher := &stubSearcher{}

	assert.Equal(t, fiber.StatusForbidden, adlStatus(t, searcher, denyAuthorizer{}, ""))
	assert.Nil(t, searcher.got, "a denied caller must NOT reach the log")
}

func TestADLHandler_Search_InvalidCriteria(t *testing.T) {
	t.Parallel()

	for _, query := range []string{"?recent=soon", "?policies=x"} {
		assert.Equal(t, fiber.StatusBadRequest, adlStatus(t, &stubSearcher{}, settingsAuth(t), query), query)
		// the denial comes before the parse error.
		assert.Equal(t, fiber.StatusForbidden, adlStatus(t, &stubSearcher{}, denyAuthorizer{}, query), query)
	}
}

// failingSearcher is a search.Searcher whose store is down.
type failingSearcher struct{}

func (failingSearcher) Search(context.Context, *search.Criteria) (authlog.AuthlogEntries, error) {
	return nil, errors.New("dial tcp 10.0.0.7:5432: connection refused")
}

func TestADLHandler_Search_FailureHidesTheCause(t *testing.T) {
	t.Parallel()

	logger := settingsLogger()
	srv := fiber.New(fiber.Config{ErrorHandler: server.ErrorHandler(logger)})
	srv.Get("/v1/adl/entries", NewADLHandler(logger, failingSearcher{}, settingsAuth(t)).Search)

	resp, err := srv.Test(httptest.NewRequest(fiber.MethodGet, "/v1/adl/entries", http.NoBody), 60000)
	require.NoError(t, err)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	assert.NotContains(t, string(body), "10.0.0.7", "a store error must NOT reach the caller")
}
