package pip

import (
	"log/slog"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestPip_CollectAttributesFromRequest(t *testing.T) {
	emptyHTTP := models.NewAttribute("http", map[string]any{})
	emptyHeaders := models.NewAttribute("headers", map[string]string{})

	testCases := []struct {
		name    string
		level   slog.Level
		req     models.Request
		attr    models.AttributeSet
		wantLog int
		wantURI string
		want    models.AttributeSet
	}{
		{
			name: "empty",
			req:  models.Request{},
			attr: models.NewAttributeSet(),
			want: models.NewAttributeSet(emptyHTTP, emptyHeaders),
		},
		{
			name: "method",
			req:  models.Request{Method: "POST"},
			attr: models.NewAttributeSet(),
			want: models.NewAttributeSet(
				emptyHeaders,
				models.NewAttributeSet(
					models.NewAttribute("http", map[string]any{"method": "POST"}),
				),
			),
		},
		{
			name: "url",
			req: models.Request{URL: &url.URL{
				Scheme:   "https://",
				Host:     "www.disney.land",
				Path:     "/donald/duck",
				RawQuery: "x=y&q=www",
			}},
			attr: models.NewAttributeSet(),
			want: models.NewAttributeSet(
				emptyHeaders,
				models.NewAttributeSet(
					models.NewAttribute("http", map[string]any{
						"scheme":     "https",
						"host":       "www.disney.land",
						"path":       "/donald/duck",
						"path-parts": []string{"donald", "duck"},
						"query":      map[string]string{"x": "y", "q": "www"},
					}),
				),
			),
		},
		{
			name: "headers",
			req:  models.Request{Headers: map[string][]string{"Content-Type": {"text/json"}, "hello": {"kitties", "world"}}},
			attr: models.NewAttributeSet(),
			want: models.NewAttributeSet(
				emptyHTTP,
				models.NewAttributeSet(models.NewAttribute("headers", map[string]string{"hello": "kitties,world"})),
				models.NewAttributeSet(models.NewAttribute("content-type", "text/json")),
			),
		},
		{
			name: "attributes",
			req:  models.Request{Attributes: map[string]any{"hello": "world", "int": 4567}},
			attr: models.NewAttributeSet(),
			want: models.NewAttributeSet(
				emptyHTTP,
				emptyHeaders,
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
					Scheme:   "https://",
					Host:     "www.disney.land",
					Path:     "/donald/duck",
					RawQuery: "x=y&q=www",
				},
				Headers:    map[string][]string{"Content-Type": {"text/json"}, "hello": {"kitties", "world"}},
				Body:       []byte(`{"float": 12.12, "bool": false, "hello": "kitty"}`),
				Attributes: map[string]any{"hello": "world", "int": 765},
			},
			attr:    models.NewAttributeSet(),
			wantLog: 1,
			want: models.NewAttributeSet(
				models.NewAttributeSet(models.NewAttribute("content-type", "text/json")),
				models.NewAttributeSet(models.NewAttribute("headers", map[string]string{"hello": "kitties,world"})),
				models.NewAttributeSet(
					models.NewAttribute("http", map[string]any{
						"method":     "POST",
						"scheme":     "https",
						"host":       "www.disney.land",
						"path":       "/donald/duck",
						"path-parts": []string{"donald", "duck"},
						"query":      map[string]string{"x": "y", "q": "www"},
					}),
				),
				models.NewAttributeSet(
					models.NewAttribute("hello", "world"),
					models.NewAttribute("int", 765),
					models.NewAttribute("body", map[string]any{"float": 12.12, "bool": false, "hello": "kitty"}),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(tc.level)
			p := &pip{logger: slog.New(h), attributes: tc.attr, newAttributes: models.NewAttributeSet}

			got, newURI := p.CollectAttributesFromRequest(&tc.req)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantURI, newURI)

			assert.Equal(t, tc.wantLog, h.Count())

			got.RemoveAttribute("request-time")

			tc.want.IterateAttributes(func(attr models.Attribute) {
				v2 := got.GetAttributeValue(attr.Key())
				assert.EqualValues(t, attr.Value(), v2)
			})

			got.IterateAttributes(func(attr models.Attribute) {
				v2 := tc.want.GetAttributeValue(attr.Key())
				assert.EqualValues(t, attr.Value(), v2)
			})
		})
	}
}
