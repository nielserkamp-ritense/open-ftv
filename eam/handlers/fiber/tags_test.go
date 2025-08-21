package fiber

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
)

func TestNewTagsHandler(t *testing.T) {
	t.Parallel()

	t.Run("new tags handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p := pap.New(ctx, logger)
		require.NotNil(t, p)

		th := NewTagsHandler(logger, p, nil)
		require.NotNil(t, th)
	})
}

func TestTagsHandler_GetPolicies(t *testing.T) {
	t.Parallel()

	t.Run("get tags", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		db, err := pgxmock.NewPool()
		require.NoError(t, err)

		db.ExpectBegin()
		db.ExpectQuery("SELECT tag,title,description,created,created_by,updated,updated_by FROM tag").
			WillReturnRows(pgxmock.NewRows([]string{"tag", "title", "description", "created", "created_by", "updated", "updated_by"}).
				AddRow("t1", "tag1", "", "", "", "", "").
				AddRow("t2", "tag2", "", "", "", "", "").
				AddRow("t3", "tag3", "", "", "", "", "").
				AddRow("t4", "tag4", "", "", "", "", ""))
		db.ExpectCommit()

		pool, err2 := postgresql.NewWithPool(ctx, "postgres://localhost:5432/table", time.Minute, 3, db)
		require.NoError(t, err2)

		p := pap.New(ctx, logger, pap.WithPgPool(pool))

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		ph := NewTagsHandler(logger, p, auth)
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/tags", ph.GetTags)

		req := httptest.NewRequest(fiber.MethodGet, "/v1/tags", nil)
		resp, err3 := srv.Test(req, 60000)
		require.NoError(t, err3)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, PoliciesVersion, resp.Header.Get(HeaderVersion))
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, b)

		var list []*policies.Tag
		err = json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.Equal(t, 4, len(list))
	})
}
