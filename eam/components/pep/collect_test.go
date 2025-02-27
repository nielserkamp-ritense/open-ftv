package pep

import (
	"log/slog"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestPip_PARCFromRequest(t *testing.T) {
	emptyHTTP := models.NewAttribute("http", map[string]any{})

	testCases := []struct {
		name    string
		level   slog.Level
		req     models.Request
		wantLog int
		want    models.AttributeSet
	}{
		{
			name: "empty",
			req:  models.Request{},
			want: models.NewAttributeSet(emptyHTTP),
		},
		{
			name: "method",
			req:  models.Request{Method: "POST"},
			want: models.NewAttributeSet(
				models.NewAttributeSet(
					models.NewAttribute("http", map[string]any{"method": "POST"}),
					models.NewAttribute("action", "name::can_update"),
				),
			),
		},
		{
			name: "url",
			req: models.Request{URL: &url.URL{
				Scheme:   "https",
				Host:     "www.disney.land",
				Path:     "/donald/duck",
				RawQuery: "x=y&q=www",
			}},
			want: models.NewAttributeSet(
				models.NewAttribute("http", map[string]any{
					"scheme":     "https",
					"host":       "www.disney.land",
					"path":       "/donald/duck",
					"path-parts": []string{"donald", "duck"},
					"query":      map[string]string{"x": "y", "q": "www"},
				}),
				models.NewAttribute("resource", "service::https://www.disney.land/donald/duck?x=y&q=www"),
			),
		},
		{
			name: "headers",
			req:  models.Request{Headers: map[string][]string{"Content-Type": {"text/json"}, "hello": {"kitties", "world"}}},
			want: models.NewAttributeSet(
				emptyHTTP,
				models.NewAttributeSet(models.NewAttribute("headers", map[string]string{"hello": "kitties,world"})),
				models.NewAttributeSet(models.NewAttribute("content-type", "text/json")),
			),
		},
		{
			name: "attributes",
			req:  models.Request{PARC: models.PARC{Context: models.NewAttributeSet(map[string]any{"hello": "world", "int": 4567})}},
			want: models.NewAttributeSet(
				emptyHTTP,
				models.NewAttributeSet(
					models.NewAttribute("hello", "world"),
					models.NewAttribute("int", 4567),
				),
			),
		},
		{
			name:  "all with log",
			level: slog.LevelDebug,
			req: models.Request{
				Method: "POST",
				URL: &url.URL{
					Scheme:   "https",
					Host:     "www.disney.land",
					Path:     "/donald/duck",
					RawQuery: "x=y&q=www",
				},
				Headers: map[string][]string{"Content-Type": {"text/json"}, "hello": {"kitties", "world"}},
				Body:    []byte(`{"float": 12.12, "bool": false, "hello": "kitty"}`),
				PARC:    models.PARC{Context: models.NewAttributeSet(map[string]any{"hello": "world", "int": 765})},
			},
			wantLog: 1,
			want: models.NewAttributeSet(
				models.NewAttribute("content-type", "text/json"),
				models.NewAttribute("headers", map[string]string{"hello": "kitties,world"}),
				models.NewAttribute("http", map[string]any{
					"method":     "POST",
					"scheme":     "https",
					"host":       "www.disney.land",
					"path":       "/donald/duck",
					"path-parts": []string{"donald", "duck"},
					"query":      map[string]string{"x": "y", "q": "www"},
				}),
				models.NewAttribute("hello", "world"),
				models.NewAttribute("int", 765),
				models.NewAttribute("body", map[string]any{"float": 12.12, "bool": false, "hello": "kitty"}),
				models.NewAttribute("action", "name::can_update"),
				models.NewAttribute("resource", "service::https://www.disney.land/donald/duck?x=y&q=www"),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(tc.level)
			p := &pep{logger: slog.New(h)}

			e := models.NewEntitySet()

			uid, _ := uuid.NewUUID()
			tc.req.UID = &uid

			got := p.PARCFromRequest(&tc.req, e)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantLog, h.Count())

			got.Context.RemoveAttribute("time")

			tc.want.IterateAttributes(func(attr models.Attribute) {
				v2 := got.Context.GetAttributeValue(attr.Key())
				assert.EqualValues(t, attr.Value(), v2)
			})

			got.Context.IterateAttributes(func(attr models.Attribute) {
				v2 := tc.want.GetAttributeValue(attr.Key())
				assert.EqualValues(t, attr.Value(), v2)
			})
		})
	}
}

func TestPip_PARCFromHTTP(t *testing.T) {
	emptyHTTP := models.NewAttribute("http", map[string]any{})

	testCases := []struct {
		name    string
		level   slog.Level
		req     models.HTTPRequest
		attrs   models.AttributeSet
		wantLog int
		want    models.AttributeSet
	}{
		{
			name: "empty",
			req:  models.HTTPRequest{},
			want: models.NewAttributeSet(emptyHTTP),
		},
		{
			name: "method",
			req:  models.HTTPRequest{Method: "POST"},
			want: models.NewAttributeSet(
				models.NewAttribute("http", map[string]any{"method": "POST"}),
				models.NewAttribute("action", "name::can_update"),
			),
		},
		{
			name: "url",
			req: models.HTTPRequest{URL: &url.URL{
				Scheme:   "https",
				Host:     "www.disney.land",
				Path:     "/donald/duck",
				RawQuery: "x=y&q=www",
			}},
			want: models.NewAttributeSet(
				models.NewAttribute("http", map[string]any{
					"scheme":     "https",
					"host":       "www.disney.land",
					"path":       "/donald/duck",
					"path-parts": []string{"donald", "duck"},
					"query":      map[string]string{"x": "y", "q": "www"},
				}),
				models.NewAttribute("resource", "service::https://www.disney.land/donald/duck?x=y&q=www"),
			),
		},
		{
			name: "headers",
			req:  models.HTTPRequest{Headers: map[string][]string{"Content-Type": {"text/json"}, "hello": {"kitties", "world"}}},
			want: models.NewAttributeSet(
				emptyHTTP,
				models.NewAttribute("headers", map[string]string{"hello": "kitties,world"}),
				models.NewAttribute("content-type", "text/json"),
			),
		},
		{
			name:  "attributes",
			req:   models.HTTPRequest{},
			attrs: models.NewAttributeSet(models.NewAttribute("hello", "world"), models.NewAttribute("int", 4567)),
			want: models.NewAttributeSet(
				emptyHTTP,
				models.NewAttribute("hello", "world"),
				models.NewAttribute("int", 4567),
			),
		},
		{
			name:  "all with log",
			level: slog.LevelDebug,
			req: models.HTTPRequest{
				Method: "POST",
				URL: &url.URL{
					Scheme:   "https",
					Host:     "www.disney.land",
					Path:     "/donald/duck",
					RawQuery: "x=y&q=www",
				},
				Headers: map[string][]string{"Content-Type": {"text/json"}, "hello": {"kitties", "world"}},
				Body:    []byte(`{"float": 12.12, "bool": false, "hello": "kitty"}`),
			},
			attrs:   models.NewAttributeSet(models.NewAttribute("int", 765), models.NewAttribute("hello", "world")),
			wantLog: 1,
			want: models.NewAttributeSet(
				models.NewAttribute("content-type", "text/json"),
				models.NewAttribute("headers", map[string]string{"hello": "kitties,world"}),
				models.NewAttribute("http", map[string]any{
					"method":     "POST",
					"scheme":     "https",
					"host":       "www.disney.land",
					"path":       "/donald/duck",
					"path-parts": []string{"donald", "duck"},
					"query":      map[string]string{"x": "y", "q": "www"},
				}),
				models.NewAttribute("hello", "world"),
				models.NewAttribute("int", 765),
				models.NewAttribute("body", map[string]any{"float": 12.12, "bool": false, "hello": "kitty"}),
				models.NewAttribute("action", "name::can_update"),
				models.NewAttribute("resource", "service::https://www.disney.land/donald/duck?x=y&q=www"),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(tc.level)
			p := &pep{logger: slog.New(h)}

			e := models.NewEntitySet()

			uid, _ := uuid.NewUUID()

			got := p.PARCFromHTTP(uid, &tc.req, tc.attrs, e)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantLog, h.Count())

			got.Context.RemoveAttribute("time")

			tc.want.IterateAttributes(func(attr models.Attribute) {
				v2 := got.Context.GetAttributeValue(attr.Key())
				assert.EqualValues(t, attr.Value(), v2)
			})

			got.Context.IterateAttributes(func(attr models.Attribute) {
				v2 := tc.want.GetAttributeValue(attr.Key())
				assert.EqualValues(t, attr.Value(), v2)
			})
		})
	}
}
