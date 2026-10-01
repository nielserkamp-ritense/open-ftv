package fiber

import (
	"bytes"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	opa_embedded "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/opa-embedded"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestAuthHandler_AuthZENEvaluations(t *testing.T) {
	t.Parallel()

	const (
		denied  = `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"GET"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
		allowed = `{"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"POST"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`
	)

	testCases := []struct {
		name       string
		store      string
		in         string
		wantStatus int
		want       string
	}{
		{
			name:       "execute all by default",
			store:      "../../../testdata/unittest/opa",
			in:         `{"evaluations":[` + denied + `,` + allowed + `]}`,
			wantStatus: fiber.StatusOK,
			want:       `{"evaluations":[{"decision":false},{"decision":true}]}`,
		},
		{
			name:       "deny on first deny",
			store:      "../../../testdata/unittest/opa",
			in:         `{"evaluations":[` + denied + `,` + allowed + `],"options":{"evaluations_semantic":"deny_on_first_deny"}}`,
			wantStatus: fiber.StatusOK,
			want:       `{"evaluations":[{"decision":false}]}`,
		},
		{
			name:       "no evaluations is a single evaluation",
			store:      "../../../testdata/unittest/opa",
			in:         allowed,
			wantStatus: fiber.StatusOK,
			want:       `{"decision":true}`,
		},
		{
			name:       "empty evaluations is a single evaluation",
			store:      "../../../testdata/unittest/opa",
			in:         `{"evaluations":[],"subject":{"type":"doelbinding","id":"subsidies"},"action":{"name":"can_read","properties":{"method":"GET"}},"resource":{"type":"service","id":"https://inway-fsc-nlx-inway:443/brp-personen"}}`,
			wantStatus: fiber.StatusOK,
			want:       `{"decision":false}`,
		},
		{
			name:       "no evaluations and no subject",
			store:      "../../../testdata/unittest/opa",
			in:         `{"action":{"name":"can_read"},"resource":{"type":"service","id":"x"}}`,
			wantStatus: fiber.StatusBadRequest,
			want:       `{"detail":"invalid subject","status":400,"title":"Bad Request"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog2.NewDummyHandler(slog.LevelDebug))

			ep := pep.New(t.Context(), logger)

			ip, err := pip.New(t.Context(), logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore("../../../testdata/pip", true))
			require.NoError(t, err)

			ap, err := pap.New(t.Context(), logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("rego"), pap.WithFileStore(tc.store, true))
			require.NoError(t, err)

			controller := opa_embedded.NewController(pdp.WithContext(t.Context()), pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
			require.NotNil(t, controller)

			auth := NewAuthHandlerZEN(logger, nil, controller, "http://localhost/v1").WithEvaluations()

			app := fiber.New()
			app.Post("/v1/authzen/evaluations", auth.Evaluations)

			req := httptest.NewRequest(fiber.MethodPost, "/v1/authzen/evaluations", bytes.NewReader([]byte(tc.in)))

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer resp.Body.Close()

			assert.Equal(t, tc.wantStatus, resp.StatusCode)

			data, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, tc.want, string(data))
		})
	}
}
