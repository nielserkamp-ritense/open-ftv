package openfga_embedded

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage"
	fgamemory "github.com/openfga/openfga/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handlers "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/handlers/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

var errDatastore = errors.New("datastore unavailable")

// failingDatastore keeps models and tuples, but fails every tuple read a check does, like a datastore that
// went away.
type failingDatastore struct {
	storage.OpenFGADatastore
}

func (failingDatastore) ReadUserTuple(context.Context, string, storage.ReadUserTupleFilter, storage.ReadUserTupleOptions) (*openfgav1.Tuple, error) {
	return nil, errDatastore
}

func (failingDatastore) ReadUsersetTuples(context.Context, string, storage.ReadUsersetTuplesFilter, storage.ReadUsersetTuplesOptions) (storage.TupleIterator, error) {
	return nil, errDatastore
}

func (failingDatastore) Read(context.Context, string, storage.ReadFilter, storage.ReadOptions) (storage.TupleIterator, error) {
	return nil, errDatastore
}

func TestAuthZEN_InternalError(t *testing.T) {
	const item = `"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_update"},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}`

	testCases := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		want       string
	}{
		{
			name:       "evaluation",
			path:       "/evaluation",
			body:       `{` + item + `}`,
			wantStatus: http.StatusInternalServerError,
			want:       `{"detail":"AuthZEN evaluation failed","status":500,"title":"Internal Server Error"}`,
		},
		{
			name:       "evaluations",
			path:       "/evaluations",
			body:       `{"evaluations":[{` + item + `}]}`,
			wantStatus: http.StatusOK,
			want:       `{"evaluations":[{"decision":false,"context":{"error":{"status":500,"message":"datastore unavailable"}}}]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

			ip, err := pip.New(t.Context(), logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)

			c, ok := NewController(pdp.WithPIP(ip), pdp.WithLogger(logger)).(*controller)
			require.True(t, ok)

			c.engine, _ = server.NewServerWithOpts(server.WithDatastore(failingDatastore{fgamemory.New()}), server.WithLogger(newZapper(logger)))
			require.NotNil(t, c.engine)

			ap, err := pap.New(t.Context(), logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("openfga"), pap.WithFileStore("../../../testdata/unittest/openfga", true))
			require.NoError(t, err)

			c.PAP = ap
			ap.AddEventSink(c)
			ap.LoadFiles()

			zen := handlers.NewAuthHandlerZEN(logger, nil, c, "").WithEvaluations()

			app := fiber.New()
			app.Post("/evaluation", zen.Evaluation)
			app.Post("/evaluations", zen.Evaluations)

			req := httptest.NewRequest(http.MethodPost, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.wantStatus, resp.StatusCode)
			assert.JSONEq(t, tc.want, string(body))
		})
	}
}
