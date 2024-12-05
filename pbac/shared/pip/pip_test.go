package pip

import (
	"log/slog"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control"
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
		wantAttributes standards.AttributeSet
		wantEntities   standards.EntitySet
	}{
		{
			name:           "no store",
			recurse:        true,
			wantLog:        1,
			wantAttributes: standards.NewAttributeSet(),
			wantEntities:   standards.NewEntitySet(),
		},
		{
			name:           "invalid store",
			path:           "/not/a/valid/path",
			recurse:        true,
			wantLog:        1,
			wantAttributes: standards.NewAttributeSet(),
			wantEntities:   standards.NewEntitySet(),
		},
		{
			name:    "with store, no recurse",
			level:   slog.LevelDebug,
			path:    "../../../testdata/unittest/pip",
			wantLog: 1,
			wantAttributes: standards.NewAttributeSet(
				standards.NewAttribute("maandag", 1),
				standards.NewAttribute("dinsdag", 2),
				standards.NewAttribute("woensdag", 3),
				standards.NewAttribute("donderdag", 4),
				standards.NewAttribute("vrijdag", 5),
			),
			wantEntities: standards.NewEntitySet(
				standards.NewEntity("app", "app1", standards.NewAttributeSet(
					standards.NewAttribute("code", "app1"),
					standards.NewAttribute("name", "App-1"),
				)),
				standards.NewEntity("app", "app2", standards.NewAttributeSet(
					standards.NewAttribute("code", "app2"),
					standards.NewAttribute("name", "App-2"),
				)),
				standards.NewEntity("app", "app3", standards.NewAttributeSet(
					standards.NewAttribute("code", "app3"),
					standards.NewAttribute("name", "App-3"),
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
				tc.wantEntities.IterateEntities(func(e1 standards.Entity) {
					e2 := p2.entities.GetEntity(e1.UID())
					require.NotNil(t, e2)
					assert.EqualValues(t, e1, e2)
				})

				p2.entities.IterateEntities(func(e1 standards.Entity) {
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
		p := &pip{attributes: standards.NewAttributeSet()}
		require.NotNil(t, p)

		p.AddAttribute("hello", "world")
		p.AddAttribute("int", "987")
		p.AddAttribute("float", "987.789")

		assert.Equal(t, "world", p.GetAttribute("hello"))
		assert.Nil(t, p.GetAttribute("bool"))

		p2 := &pip{attributes: standards.NewAttributeSet(standards.NewAttribute("hello", "world2"), &standards.Attribute{Key: "bool", Value: true})}
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
		p := &pip{entities: standards.NewEntitySet()}
		require.NotNil(t, p)

		p.AddEntity(standards.NewEntity("x", "y", standards.NewAttributeSet()))
		p.AddEntity(standards.NewEntity("x", "z", standards.NewAttributeSet()))

		p.MergeEntities(
			standards.NewEntitySet(
				standards.NewEntity("q", "x", standards.NewAttributeSet()),
				standards.NewEntity("q", "y", standards.NewAttributeSet()),
				standards.NewEntity("q", "z", standards.NewAttributeSet()),
			),
		)

		var count int
		p.IterateEntities(func(entity standards.Entity) {
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
		p.IterateEntities(func(entity standards.Entity) {
			count++
		})
		assert.Equal(t, 3, count)

		e = p.GetEntity("q::y")
		require.Nil(t, e)
	})
}

func TestPip_CollectAttributesFromRequest(t *testing.T) {
	emptyHTTP := standards.NewAttribute("http", map[string]any{})
	emptyHeaders := standards.NewAttribute("headers", map[string]string{})

	testCases := []struct {
		name    string
		level   slog.Level
		req     control.Request
		attr    standards.AttributeSet
		wantLog int
		wantURI string
		want    standards.AttributeSet
	}{
		{
			name: "empty",
			req:  control.Request{},
			attr: standards.NewAttributeSet(),
			want: standards.NewAttributeSet(emptyHTTP, emptyHeaders),
		},
		{
			name: "method",
			req:  control.Request{Method: "POST"},
			attr: standards.NewAttributeSet(),
			want: standards.NewAttributeSet(
				emptyHeaders,
				standards.NewAttributeSet(
					standards.NewAttribute("http", map[string]any{"method": "POST"}),
				),
			),
		},
		{
			name: "url",
			req: control.Request{URL: &url.URL{
				Scheme:   "https://",
				Host:     "www.disney.land",
				Path:     "/donald/duck",
				RawQuery: "x=y&q=www",
			}},
			attr: standards.NewAttributeSet(),
			want: standards.NewAttributeSet(
				emptyHeaders,
				standards.NewAttributeSet(
					standards.NewAttribute("http", map[string]any{
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
			req:  control.Request{Headers: map[string][]string{"Content-Type": {"text/json"}, "hello": {"kitties", "world"}}},
			attr: standards.NewAttributeSet(),
			want: standards.NewAttributeSet(
				emptyHTTP,
				standards.NewAttributeSet(standards.NewAttribute("headers", map[string]string{"hello": "kitties,world"})),
				standards.NewAttributeSet(standards.NewAttribute("content-type", "text/json")),
			),
		},
		{
			name: "attributes",
			req:  control.Request{Attributes: map[string]any{"hello": "world", "int": 4567}},
			attr: standards.NewAttributeSet(),
			want: standards.NewAttributeSet(
				emptyHTTP,
				emptyHeaders,
				standards.NewAttributeSet(
					standards.NewAttribute("hello", "world"),
					standards.NewAttribute("int", 4567),
				),
			),
		},
		{
			name:  "all with log",
			level: slog.LevelDebug,
			req: control.Request{
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
			attr:    standards.NewAttributeSet(),
			wantLog: 1,
			want: standards.NewAttributeSet(
				standards.NewAttributeSet(standards.NewAttribute("content-type", "text/json")),
				standards.NewAttributeSet(standards.NewAttribute("headers", map[string]string{"hello": "kitties,world"})),
				standards.NewAttributeSet(
					standards.NewAttribute("http", map[string]any{
						"method":     "POST",
						"scheme":     "https",
						"host":       "www.disney.land",
						"path":       "/donald/duck",
						"path-parts": []string{"donald", "duck"},
						"query":      map[string]string{"x": "y", "q": "www"},
					}),
				),
				standards.NewAttributeSet(
					standards.NewAttribute("hello", "world"),
					standards.NewAttribute("int", 765),
					standards.NewAttribute("body", map[string]any{"float": 12.12, "bool": false, "hello": "kitty"}),
				),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := util.NewDummyHandler(tc.level)
			p := &pip{logger: slog.New(h), attributes: tc.attr, newAttributes: standards.NewAttributeSet}

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
