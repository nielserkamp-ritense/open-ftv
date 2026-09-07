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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNewLanguagesHandler(t *testing.T) {
	t.Parallel()

	t.Run("new languages handler", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		p, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""))
		require.NoError(t, err)
		require.NotNil(t, p)

		lh := NewLanguagesHandler(logger, p, nil)
		require.NotNil(t, lh)
	})
}

func TestLanguagesHandler_GetLanguages(t *testing.T) {
	t.Parallel()

	t.Run("get languages", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		auth := authorization.New(authorization.NoAuth(), authorization.WithAuthenticator(authentication.NewDummy()))
		require.NotNil(t, auth)

		db, err := pgxmock.NewPool()
		require.NoError(t, err)

		db.ExpectBegin()
		db.ExpectQuery("SELECT language,title,description,created,created_by,updated,updated_by FROM language").
			WillReturnRows(pgxmock.NewRows([]string{"language", "title", "description", "created", "created_by", "updated", "updated_by"}).
				AddRow("rego", "Rego", "", "", "", "", "").
				AddRow("cedar", "Cedar", "", "", "", "", "").
				AddRow("cerbos", "Cerbos", "", "", "", "", "").
				AddRow("openfga", "OpenFGA", "", "", "", "", ""))
		db.ExpectCommit()

		pool, err2 := postgresql.NewWithPool(ctx, "postgres://localhost:5432/table", time.Minute, 3, db)
		require.NoError(t, err2)

		p, err := pap.New(ctx, logger, pap.WithPgPool(pool))
		require.NoError(t, err)
		require.NotNil(t, p)

		ph := NewLanguagesHandler(logger, p, auth)
		require.NotNil(t, ph)

		srv := fiber.New()
		srv.Get("/v1/languages", ph.GetLanguages)

		req := httptest.NewRequest(fiber.MethodGet, "/v1/languages", nil)
		resp, err3 := srv.Test(req, 60000)
		require.NoError(t, err3)
		require.NotNil(t, resp)
		defer resp.Body.Close()

		assert.Equal(t, PoliciesVersion, resp.Header.Get(HeaderVersion))
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		b, err4 := io.ReadAll(resp.Body)
		require.NoError(t, err4)
		require.NotNil(t, b)

		var list policies.Languages
		err = json.Unmarshal(b, &list)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(list), 4)
	})
}
