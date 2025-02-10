package cedar

import (
	"log/slog"
	"net/url"
	"testing"
	"time"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestController_Authorize(t *testing.T) {
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
		req      components.Request
		wantErr  bool
		wantLog  int
		want     components.Response
	}{
		{
			name:     "bad request",
			store1:   "../../../../testdata/pip",
			recurse1: true,
			store2:   "../../../../testdata/unittest/cedar",
			recurse2: true,
			req: components.Request{
				UID:         &uid,
				URL:         u1,
				Method:      "GET",
				RequestTime: &now,
				Headers:     map[string][]string{},
				Body:        []byte(""),
			},
			wantLog: 4,
			want:    components.Response{Message: "not authorized", Attributes: map[string]any{"diagnostic": types.Diagnostic{}}},
		},
		{
			name:     "good request",
			store1:   "../../../../testdata/pip",
			recurse1: true,
			store2:   "../../../../testdata/unittest/cedar",
			recurse2: true,
			req: components.Request{
				UID:         &uid,
				URL:         u2,
				Method:      "POST",
				RequestTime: &now,
				Headers:     map[string][]string{"Content-Type": {"application/json"}},
				Body:        b1,
			},
			wantLog: 3,
			want:    components.Response{Allowed: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(pip.Config{
				Store:         tc.store1,
				Recurse:       tc.recurse1,
				Logger:        logger,
				NewAttributes: models.NewAttributeSet,
				NewEntities:   models.NewEntitySet,
			})
			require.NotNil(t, p)

			c := NewController(pdp.WithPIP(p), pdp.WithStore(tc.store2, tc.recurse2), pdp.WithLogger(logger))
			require.NotNil(t, c)

			h.Clear()

			got, err := c.Authorize(&tc.req)

			assert.Equal(t, tc.wantLog, h.Count())

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
