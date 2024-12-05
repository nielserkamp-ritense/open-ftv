package pip

import (
	"log/slog"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestValidPath(t *testing.T) {
	testCases := []struct {
		name string
		path string
		want bool
	}{
		{name: "empty"},
		{name: "dot", path: ".", want: true},
		{name: "dot dot", path: "..", want: true},
		{name: "invalid", path: "/not/really/a/valid/path"},
		{name: "valid", path: "/usr/sbin", want: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := validPath(tc.path)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNew(t *testing.T) {
	testCases := []struct {
		name           string
		level          slog.Level
		path           string
		recurse        bool
		wantLog        int
		wantAttributes models.AttributeSet
		wantEntities   models.EntitySet
	}{
		{
			name:           "no store",
			recurse:        true,
			wantLog:        1,
			wantAttributes: models.NewAttributeSet(),
			wantEntities:   models.NewEntitySet(),
		},
		{
			name:           "invalid store",
			path:           "/not/a/valid/path",
			recurse:        true,
			wantLog:        1,
			wantAttributes: models.NewAttributeSet(),
			wantEntities:   models.NewEntitySet(),
		},
		{
			name:    "with store, no recurse",
			level:   slog.LevelDebug,
			path:    "../../../testdata/unittest/pip",
			wantLog: 1,
			wantAttributes: models.NewAttributeSet(
				models.NewAttribute("maandag", 1),
				models.NewAttribute("dinsdag", 2),
				models.NewAttribute("woensdag", 3),
				models.NewAttribute("donderdag", 4),
				models.NewAttribute("vrijdag", 5),
			),
			wantEntities: models.NewEntitySet(
				models.NewEntity("app", "app1", models.NewAttributeSet(
					models.NewAttribute("code", "app1"),
					models.NewAttribute("name", "App-1"),
				)),
				models.NewEntity("app", "app2", models.NewAttributeSet(
					models.NewAttribute("code", "app2"),
					models.NewAttribute("name", "App-2"),
				)),
				models.NewEntity("app", "app3", models.NewAttributeSet(
					models.NewAttribute("code", "app3"),
					models.NewAttribute("name", "App-3"),
				), "app::app1", "app::app2",
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(tc.level)

			p := New(nil, tc.path, tc.recurse, slog.New(h), nil, nil)
			require.NotNil(t, p)

			p2, ok := p.(*pip)
			require.True(t, ok)
			require.NotNil(t, p2)

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.wantAttributes != nil {
				tc.wantAttributes.IterateAttributes(func(key string, v1 any) {
					v2 := p2.attributes.GetAttribute(key)
					assert.EqualValues(t, v1, v2)
				})

				p2.attributes.IterateAttributes(func(key string, v1 any) {
					v2 := tc.wantAttributes.GetAttribute(key)
					assert.EqualValues(t, v1, v2)
				})
			}

			if tc.wantEntities != nil {
				tc.wantEntities.IterateEntities(func(e1 models.Entity) {
					e2 := p2.entities.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.EqualValues(t, e1, e2)
				})

				p2.entities.IterateEntities(func(e1 models.Entity) {
					e2 := tc.wantEntities.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.EqualValues(t, e1, e2)
				})
			}
		})
	}
}

func TestPIP_Attributes(t *testing.T) {
	t.Run("pip as AttributeSet", func(t *testing.T) {
		p := &pip{attributes: models.NewAttributeSet()}
		require.NotNil(t, p)

		p.AddAttribute("hello", "world")
		p.AddAttribute("int", "987")
		p.AddAttribute("float", "987.789")

		assert.Equal(t, "world", p.GetAttribute("hello"))
		assert.Nil(t, p.GetAttribute("bool"))

		p2 := &pip{attributes: models.NewAttributeSet(models.NewAttribute("hello", "world2"), &models.Attribute{Key: "bool", Value: true})}
		p.MergeAttributes(p2)

		assert.Equal(t, "world2", p.GetAttribute("hello"))
		assert.Equal(t, true, p.GetAttribute("bool"))

		p.RemoveAttribute("bool")
		p.RemoveAttribute("int")
		assert.Nil(t, p.GetAttribute("bool"))
		assert.Nil(t, p.GetAttribute("int"))

		var count int
		p.IterateAttributes(func(string, any) {
			count++
		})
		assert.Equal(t, 2, count)
	})
}

func TestPIP_Entities(t *testing.T) {
	t.Run("pip as EntitySet", func(t *testing.T) {
		p := &pip{entities: models.NewEntitySet()}
		require.NotNil(t, p)

		p.AddEntity(models.NewEntity("x", "y", models.NewAttributeSet()))
		p.AddEntity(models.NewEntity("x", "z", models.NewAttributeSet()))

		p.MergeEntities(
			models.NewEntitySet(
				models.NewEntity("q", "x", models.NewAttributeSet()),
				models.NewEntity("q", "y", models.NewAttributeSet()),
				models.NewEntity("q", "z", models.NewAttributeSet()),
			),
		)

		var count int
		p.IterateEntities(func(entity models.Entity) {
			count++
		})
		assert.Equal(t, 5, count)

		e := p.GetEntity("x::y")
		require.NotNil(t, e)

		e = p.GetEntity("x::x")
		require.Nil(t, e)

		p.RemoveEntity("q::y")
		p.RemoveEntity("q::x")

		count = 0
		p.IterateEntities(func(entity models.Entity) {
			count++
		})
		assert.Equal(t, 3, count)

		e = p.GetEntity("q::y")
		require.Nil(t, e)
	})
}

func TestPip_CollectAttributesFromRequest(t *testing.T) {
	emptyHTTP := models.NewAttribute("http", map[string]any{})
	emptyHeaders := models.NewAttribute("headers", map[string]string{})

	testCases := []struct {
		name    string
		level   slog.Level
		req     shared.Request
		attr    models.AttributeSet
		wantLog int
		wantURI string
		want    models.AttributeSet
	}{
		{
			name: "empty",
			req:  shared.Request{},
			attr: models.NewAttributeSet(),
			want: models.NewAttributeSet(emptyHTTP, emptyHeaders),
		},
		{
			name: "method",
			req:  shared.Request{Method: "POST"},
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
			req: shared.Request{URL: &url.URL{
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
			req:  shared.Request{Headers: map[string][]string{"Content-Type": {"text/json"}, "hello": {"kitties", "world"}}},
			attr: models.NewAttributeSet(),
			want: models.NewAttributeSet(
				emptyHTTP,
				models.NewAttributeSet(models.NewAttribute("headers", map[string]string{"hello": "kitties,world"})),
				models.NewAttributeSet(models.NewAttribute("content-type", "text/json")),
			),
		},
		{
			name: "attributes",
			req:  shared.Request{Attributes: map[string]any{"hello": "world", "int": 4567}},
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
			req: shared.Request{
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

			tc.want.IterateAttributes(func(key string, v1 any) {
				v2 := got.GetAttribute(key)
				assert.EqualValues(t, v1, v2)
			})

			got.IterateAttributes(func(key string, v1 any) {
				v2 := tc.want.GetAttribute(key)
				assert.EqualValues(t, v1, v2)
			})
		})
	}
}
