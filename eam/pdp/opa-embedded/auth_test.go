package opa_embedded

import (
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	pip2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestController_Authorize(t *testing.T) {
	t.Parallel()

	now := time.Now()
	uid := uuid.New()
	u1, _ := url.Parse("https://x.y")
	u2, _ := url.Parse("https://inway-fsc-nlx-inway:443/brp-personen")
	b1 := []byte(`{"type":"RaadpleegMetBurgerservicenummer","burgerservicenummer":["999993653"],"fields":["naam"]}`)

	testCases := []struct {
		name     string
		store1   string
		recurse1 bool
		store2   string
		recurse2 bool
		req      models.Request
		wantErr  bool
		wantLog  int
		want     models.Response
	}{
		{
			name:     "bad request",
			store1:   "../../../testdata/pip",
			recurse1: true,
			store2:   "../../../testdata/unittest/opa",
			recurse2: true,
			req: models.Request{
				UID:         &uid,
				URL:         u1,
				Method:      "GET",
				RequestTime: &now,
				Headers: map[string][]string{
					"doelbinding": {"subsidies"},
				},
				Body: []byte(""),
			},
			wantLog: 4,
			want:    models.Response{Message: "not authorized"},
		},
		{
			name:     "good request",
			store1:   "../../../testdata/pip",
			recurse1: true,
			store2:   "../../../testdata/unittest/opa",
			recurse2: true,
			req: models.Request{
				UID:         &uid,
				URL:         u2,
				Method:      "POST",
				RequestTime: &now,
				Headers: map[string][]string{
					"doelbinding":  {"subsidies"},
					"Content-Type": {"application/json"},
				},
				Body: b1,
			},
			wantLog: 4,
			want:    models.Response{Allowed: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			ep := pep.New(nil, logger)

			ip := pip2.New(nil, logger, pip2.WithFileStore(tc.store1, tc.recurse1))
			require.NotNil(t, ip)

			ap := pap2.New(nil, logger, pap2.WithLanguage("rego"), pap2.WithFileStore(tc.store2, tc.recurse2))
			require.NotNil(t, ap)

			c := NewController(pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
			require.NotNil(t, c)

			h.Clear()

			parc := c.PEP().PARCFromRequest(&tc.req, c.PIP().GetEntity)
			require.NotNil(t, parc)

			got, err := c.Authorize(tc.req.UID.String(), parc)

			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)

			if tc.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.EqualValues(t, tc.want, *got)
			}
		})
	}
}
