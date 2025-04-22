package pep

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestRun(t *testing.T) {
	t.Parallel()

	emptyHTTP := models.NewAttribute("http", map[string]any{})
	invalidPrincipal := models.NewEntity("invalid", "invalid", models.NewAttributeSet())

	testCases := []struct {
		name    string
		debug   bool
		parc    *models.PARC
		req     *models.HTTPRequest
		uri     string
		want    *models.PARC
		wantURI string
		wantLog int
	}{
		{
			name: "nil, empty, no uri",
			parc: &models.PARC{Context: models.NewAttributeSet()},
			req:  &models.HTTPRequest{},
			want: &models.PARC{
				Principal: invalidPrincipal,
				Action:    models.NewEntity("", "", models.NewAttributeSet()),
				Resource:  models.NewEntity("", "", models.NewAttributeSet()),
				Context:   models.NewAttributeSet(emptyHTTP),
			},
		},
		{
			name: "nil, empty, uri",
			parc: &models.PARC{Context: models.NewAttributeSet()},
			req:  &models.HTTPRequest{},
			uri:  "http://localhost",
			want: &models.PARC{
				Principal: invalidPrincipal,
				Action:    models.NewEntity("", "", models.NewAttributeSet()),
				Resource:  models.NewEntity("service", "http://localhost", models.NewAttributeSet()),
				Context: models.NewAttributeSet(
					emptyHTTP,
					models.NewAttribute("resource", "service::http://localhost"),
				),
			},
			wantURI: "http://localhost",
		},
		{
			name: "nil, GET, uri",
			parc: &models.PARC{Context: models.NewAttributeSet()},
			req:  &models.HTTPRequest{Method: "GET"},
			uri:  "http://localhost",
			want: &models.PARC{
				Principal: invalidPrincipal,
				Action:    models.NewEntity("name", "can_read", models.NewAttributeSet(models.NewAttribute("method", "GET"))),
				Resource:  models.NewEntity("service", "http://localhost", models.NewAttributeSet()),
				Context: models.NewAttributeSet(
					models.NewAttribute("http", map[string]any{"method": "GET"}),
					models.NewAttribute("action", "name::can_read"),
					models.NewAttribute("resource", "service::http://localhost"),
				),
			},
			wantURI: "http://localhost",
		},
		{
			name:  "nil, GET + rvvaid, uri",
			debug: true,
			parc:  &models.PARC{Context: models.NewAttributeSet()},
			req:   &models.HTTPRequest{Method: "GET", Headers: map[string][]string{"dpl-processing-activity-id": {"abc"}}},
			uri:   "http://localhost",
			want: &models.PARC{
				Principal: models.NewEntity("activity", "abc", models.NewAttributeSet()),
				Action:    models.NewEntity("name", "can_read", models.NewAttributeSet(models.NewAttribute("method", "GET"))),
				Resource:  models.NewEntity("service", "http://localhost", models.NewAttributeSet()),
				Context: models.NewAttributeSet(
					models.NewAttribute("http", map[string]any{"method": "GET"}),
					models.NewAttribute("principal", "activity::abc"),
					models.NewAttribute("action", "name::can_read"),
					models.NewAttribute("resource", "service::http://localhost"),
					models.NewAttribute("rvva_id", "abc"),
				),
			},
			wantURI: "http://localhost",
			wantLog: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			c := &collector{
				debug:  tc.debug,
				logger: logger,
				req:    tc.req,
				parc:   tc.parc,
				newURI: tc.uri,
			}

			c.run()

			c.parc.Context.RemoveAttribute("time")

			assert.True(t, models.EntityEqual(tc.want.Principal, c.parc.Principal))
			assert.True(t, models.EntityEqual(tc.want.Action, c.parc.Action))
			assert.True(t, models.EntityEqual(tc.want.Resource, c.parc.Resource))
			assert.True(t, models.AttributesEqual(tc.want.Context, c.parc.Context))
			assert.Equal(t, tc.wantURI, c.newURI)

			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}
