package cedar_embedded

import (
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestController_Authorize(t *testing.T) {
	t.Parallel()

	now := time.Now()
	uid := uuid.New()
	u1, _ := url.Parse("https://x.y")
	u2, _ := url.Parse("https://inway-fsc-nlx-inway:443/brp-personen")
	b1 := []byte(`{"type": "RaadpleegMetBurgerservicenummer","burgerservicenummer":["999993653"],"fields":["naam"]}`)

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
			store2:   "../../../testdata/unittest/cedar",
			recurse2: true,
			req: models.Request{
				UID:         &uid,
				URL:         u1,
				Method:      "GET",
				RequestTime: &now,
				Headers:     map[string][]string{},
				Body:        []byte(""),
			},
			wantLog: 3,
			want:    models.Response{Message: "not authorized", Attributes: map[string]any{"diagnostic": types.Diagnostic{}}},
		},
		{
			name:     "good request",
			store1:   "../../../testdata/pip",
			recurse1: true,
			store2:   "../../../testdata/unittest/cedar",
			recurse2: true,
			req: models.Request{
				UID:         &uid,
				URL:         u2,
				Method:      "POST",
				RequestTime: &now,
				Headers:     map[string][]string{"Content-Type": {"application/json"}},
				Body:        b1,
			},
			wantLog: 3,
			want:    models.Response{Allowed: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			ep := pep.New(nil, logger)

			ip, err := pip.New(t.Context(), logger, pip.WithKeyValueDB(memory.New(), ""), pip.WithFileStore(tc.store1, tc.recurse1))
			require.NoError(t, err)
			require.NotNil(t, ip)

			ap, err := pap.New(t.Context(), logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"), pap.WithFileStore(tc.store2, tc.recurse2))
			require.NoError(t, err)
			require.NotNil(t, ap)

			c := NewController(pdp.WithPEP(ep), pdp.WithPIP(ip), pdp.WithPAP(ap), pdp.WithLogger(logger))
			require.NotNil(t, c)

			h.Clear()

			c2, ok := c.(*controller)
			require.True(t, ok)
			require.NotNil(t, c2)

			parc := c2.PEP.PARCFromRequest(&tc.req, c2.PIP.GetEntity)
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
