package models

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

func TestNewPolicyFromOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		p       *policies.Policy
		data    string
		wantErr bool
		wantKey string
	}{
		{
			name:    "no data",
			p:       &policies.Policy{Id: "x1"},
			wantErr: true,
		},
		{
			name:    "only id",
			p:       &policies.Policy{Id: "x1"},
			data:    "policy-1",
			wantKey: "/x1",
		},
		{
			name:    "some meta",
			p:       &policies.Policy{Id: "x2", Metadata: policies.Metadata{Title: "some title"}},
			data:    "policy-2",
			wantKey: "/x2",
		},
		{
			name: "all meta",
			p: &policies.Policy{
				Id:       "x3",
				Language: "opa",
				Metadata: policies.Metadata{
					Title:       "x3",
					Description: "description x3",
					RvvaId:      "e3",
					Url:         "https://some.site/policies/x3",
				},
			},
			data:    "policy-3",
			wantKey: "opa/x3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var d io.Reader
			if tc.data != "" {
				d = bytes.NewBufferString(tc.data)
			}

			got, err := NewPolicyFromOAS(tc.p, d)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.Equal(t, tc.wantKey, got.Key())
				assert.Equal(t, tc.p.Id, got.ID())
				assert.Equal(t, tc.p.Language, got.Language())
				assert.Equal(t, tc.p.Metadata.Title, got.Title())
				assert.Equal(t, tc.p.Metadata.Description, got.Description())
				assert.Equal(t, tc.p.Metadata.RvvaId, got.RvvaID())
				assert.Equal(t, tc.p.Metadata.Url, got.URI())

				r := got.Content()
				d, _ := io.ReadAll(r)
				assert.Equal(t, tc.data, string(d))
			}
		})
	}
}

func TestNewPolicyFromData(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		id       string
		language string
		rvvaID   string
		uri      string
		data     string
		wantErr  bool
	}{
		{
			name:    "no data",
			id:      "x0",
			wantErr: true,
		},
		{
			name: "only id",
			id:   "x1",
			data: "policy-1",
		},
		{
			name: "some meta",
			id:   "x2",
			data: "policy-2",
		},
		{
			name:     "all meta",
			id:       "x3",
			language: "opa",
			rvvaID:   "e3",
			uri:      "https://some.site/policies/x3",
			data:     "policy-3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var d io.Reader
			if tc.data != "" {
				d = bytes.NewBufferString(tc.data)
			}

			got, err := NewPolicyFromData(tc.id, tc.language, tc.rvvaID, tc.uri, d)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.Equal(t, tc.id, got.ID())
				assert.Equal(t, tc.language, got.Language())
				assert.Equal(t, tc.rvvaID, got.RvvaID())
				assert.Equal(t, tc.uri, got.URI())

				r := got.Content()
				d, _ := io.ReadAll(r)
				assert.Equal(t, tc.data, string(d))
			}
		})
	}
}

func TestNewPolicyFromStore(t *testing.T) {
	t.Parallel()

	path1 := "../../testdata/unittest/cedar/allow_post.cedar"
	path2 := "../../testdata/unittest/opa/subsidies.rego"
	path3 := "../../testdata/unittest/openfga/doelbinding.model"
	path4 := "../../testdata/unittest/opa2/bad_meta.rego"

	f1, err := os.Open(path1)
	require.NoError(t, err)
	f1.Close()

	testCases := []struct {
		name            string
		path            string
		content         io.Reader
		wantErr         bool
		wantID          string
		wantLanguage    string
		wantTitle       string
		wantDescription string
		wantTags        []string
		wantRvva        string
		wantURI         string
		wantPath        string
		wantContent     string
	}{
		{
			name:    "nil",
			path:    path1,
			wantErr: true,
		},
		{
			name:    "closed file",
			path:    path1,
			content: f1,
			wantErr: true,
		},
		{
			name:        "no metadata",
			path:        path1,
			content:     bytes.NewBufferString("some data"),
			wantID:      "allow_post.cedar",
			wantTags:    []string{},
			wantPath:    path1,
			wantContent: "some data",
		},
		{
			name:            "with yaml metadata",
			path:            path2,
			content:         bytes.NewBufferString("some data"),
			wantID:          "subsidies",
			wantDescription: "beleidsregels voor subsidies",
			wantTags:        []string{"brp", "rvva", "subsidies"},
			wantLanguage:    "rego",
			wantRvva:        "rvva1",
			wantURI:         "https://my.site/pol/x1",
			wantPath:        path2,
			wantContent:     "some data",
		},
		{
			name:            "with json metadata",
			path:            path3,
			content:         bytes.NewBufferString("some data"),
			wantID:          "doelbinding.model",
			wantTags:        []string{},
			wantLanguage:    "openfga",
			wantTitle:       "titel",
			wantDescription: "omschrijving",
			wantRvva:        "rvva2",
			wantURI:         "https://my.site/openfga/doelbinding.model",
			wantPath:        path3,
			wantContent:     "some data",
		},
		{
			name:    "with bad json",
			path:    path4,
			content: bytes.NewBufferString("some data"),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err2 := NewPolicyFromStore("", tc.path, tc.content)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, got)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, got)

				assert.Equal(t, tc.wantID, got.ID())
				assert.Equal(t, tc.wantLanguage, got.Language())
				assert.Equal(t, tc.wantTitle, got.Title())
				assert.Equal(t, tc.wantDescription, got.Description())
				assert.EqualValues(t, tc.wantTags, got.Tags())
				assert.Equal(t, tc.wantRvva, got.RvvaID())
				assert.Equal(t, tc.wantURI, got.URI())
				assert.Equal(t, tc.wantPath, got.Path())

				r := got.Content()
				d, _ := io.ReadAll(r)
				assert.Equal(t, tc.wantContent, string(d))
			}
		})
	}
}

func TestPolicy_WithTitle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		in    *Policy
		title string
	}{
		{
			name:  "no title",
			in:    &Policy{},
			title: "New title",
		},
		{
			name:  "existing title",
			in:    &Policy{title: "Old title"},
			title: "New title 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.WithTitle(tc.title)
			require.NotNil(t, got)
			assert.Equal(t, tc.title, got.Title())
		})
	}
}

func TestPolicy_WithDescription(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Policy
		desc string
	}{
		{
			name: "no description",
			in:   &Policy{},
			desc: "New description",
		},
		{
			name: "existing description",
			in:   &Policy{description: "Old description"},
			desc: "New description 2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.WithDescription(tc.desc)
			require.NotNil(t, got)
			assert.Equal(t, tc.desc, got.Description())
		})
	}
}

func TestPolicy_WithTags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		id       string
		language string
		rvva     string
		tags     []string
	}{
		{name: "single", id: "p1", language: "cedar", tags: []string{"x"}},
		{name: "few", id: "p2", language: "rego", rvva: "rvva1", tags: []string{"x", "y", "z"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewPolicyFromData(tc.id, tc.language, tc.rvva, "", bytes.NewBufferString("yo"))
			require.NoError(t, err)
			require.NotNil(t, got)

			got.WithTags(tc.tags...)
			assert.EqualValues(t, tc.tags, got.Tags())

			for i := range tc.tags {
				assert.True(t, got.HasTag(tc.tags[i]))
			}

			assert.False(t, got.HasTag("qqq"))

			if tc.rvva != "" {
				assert.True(t, got.HasTag(tc.rvva))
			}
			if tc.language != "" {
				assert.True(t, got.HasTag(tc.language))
			}
		})
	}
}

func TestPolicy_ToOAS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Policy
		data bool
		want *policies.Policy
	}{
		{
			name: "empty",
			in:   &Policy{},
			data: false,
			want: &policies.Policy{Metadata: policies.Metadata{Tags: []string{}}},
		},
		{
			name: "some data, with uri",
			in:   &Policy{language: "rego", id: "p1", title: "titel", uri: "http://localhost:8080/"},
			data: true,
			want: &policies.Policy{
				Language: "rego",
				Id:       "p1",
				Metadata: policies.Metadata{
					Title: "titel",
					Url:   "http://localhost:8080/",
					Tags:  []string{},
				},
			},
		},
		{
			name: "all data, without uri",
			in: &Policy{
				language:    "rego",
				id:          "p1",
				title:       "titel",
				description: "omschrijving",
				rvvaID:      "abc-def-gh",
				tags:        map[string]struct{}{"x": {}, "y": {}, "z": {}},
				content:     []byte("allow=true"),
			},
			data: true,
			want: &policies.Policy{
				Language: "rego",
				Id:       "p1",
				Metadata: policies.Metadata{
					Title:       "titel",
					Description: "omschrijving",
					RvvaId:      "abc-def-gh",
					Tags:        []string{"x", "y", "z"},
				},
				Data: "allow=true",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.ToOAS(tc.data)
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestPolicy_ToBundle(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *Policy
		want *policies.Policy
	}{
		{
			name: "empty",
			in:   &Policy{},
			want: &policies.Policy{},
		},
		{
			name: "some data, with uri",
			in:   &Policy{language: "rego", id: "p1", title: "titel", uri: "http://localhost:8080/"},
			want: &policies.Policy{
				Language: "rego",
				Id:       "p1",
			},
		},
		{
			name: "all data, without uri",
			in: &Policy{
				language:    "rego",
				id:          "p1",
				title:       "titel",
				description: "omschrijving",
				rvvaID:      "abc-def-gh",
				tags:        map[string]struct{}{"x": {}, "y": {}, "z": {}},
				content:     []byte("allow=true"),
			},
			want: &policies.Policy{
				Language: "rego",
				Id:       "p1",
				Data:     "allow=true",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.ToBundle()
			require.NotNil(t, got)
			assert.EqualValues(t, tc.want, got)
		})
	}
}

func TestPolicy_JSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		p    *policies.Policy
		tags []string
		data string
		want string
	}{
		{
			name: "only id",
			p:    &policies.Policy{Id: "x1"},
			data: "policy-1",
			want: `{"language":"","id":"x1","content":"cG9saWN5LTE="}`,
		},
		{
			name: "some meta",
			p:    &policies.Policy{Id: "x2"},
			data: "policy-2",
			want: `{"language":"","id":"x2","content":"cG9saWN5LTI="}`,
		},
		{
			name: "tags",
			p:    &policies.Policy{Id: "x1"},
			tags: []string{"x", "y", "z"},
			data: "policy-1",
			want: `{"language":"","id":"x1","tags":["x","y","z"],"content":"cG9saWN5LTE="}`,
		},
		{
			name: "all meta",
			p: &policies.Policy{
				Id:       "x3",
				Language: "opa",
				Metadata: policies.Metadata{
					Title:       "titel",
					Description: "omschrijving",
					RvvaId:      "e3",
					Url:         "https://some.site/policies/x3",
					Tags:        []string{"x", "y", "z"},
				},
			},
			data: "policy-3",
			want: `{"language":"opa","id":"x3","title":"titel","description":"omschrijving","tags":["x","y","z"],"rvvaID":"e3","uri":"https://some.site/policies/x3","content":"cG9saWN5LTM="}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := NewPolicyFromOAS(tc.p, bytes.NewBufferString(tc.data))
			require.NoError(t, err)
			require.NotNil(t, p)

			if len(tc.tags) > 0 {
				p.WithTags(tc.tags...)
			}

			b, err2 := json.Marshal(p)
			require.NoError(t, err2)
			require.NotNil(t, b)
			assert.Equal(t, tc.want, string(b))

			p2 := new(Policy)
			err = json.Unmarshal(b, p2)
			require.NoError(t, err)
			assert.True(t, p.Equals(p2))

			p3, err3 := NewPolicyFromData("p9", "oops", "what?", "no-uri", bytes.NewBufferString("hello"))
			require.NoError(t, err3)

			p3.WithTags("x", "y").WithDescription("nothing here")

			err3 = json.Unmarshal(b, p3)
			require.NoError(t, err3)
			assert.True(t, p.Equals(p3))
		})
	}
}

func TestPolicy_YAML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		p    *policies.Policy
		tags []string
		data string
		want string
	}{
		{
			name: "only id",
			p:    &policies.Policy{Id: "x1"},
			data: "policy-1",
			want: "language: \"\"\nid: x1\ncontent: cG9saWN5LTE=\n",
		},
		{
			name: "some meta",
			p:    &policies.Policy{Id: "x2"},
			data: "policy-2",
			want: "language: \"\"\nid: x2\ncontent: cG9saWN5LTI=\n",
		},
		{
			name: "tags",
			p:    &policies.Policy{Id: "x1"},
			tags: []string{"x", "y", "z"},
			data: "policy-1",
			want: "language: \"\"\nid: x1\ntags:\n- x\n- \"y\"\n- z\ncontent: cG9saWN5LTE=\n",
		},
		{
			name: "all meta",
			p: &policies.Policy{
				Id:       "x3",
				Language: "opa",
				Metadata: policies.Metadata{
					Title:       "titel",
					Description: "omschrijving",
					RvvaId:      "e3",
					Url:         "https://some.site/policies/x3",
					Tags:        []string{"x", "y", "z"},
				},
			},
			data: "policy-3",
			want: "language: opa\nid: x3\ntitle: titel\ndescription: omschrijving\ntags:\n- x\n- \"y\"\n- z\nrvvaID: e3\nuri: https://some.site/policies/x3\ncontent: cG9saWN5LTM=\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := NewPolicyFromOAS(tc.p, bytes.NewBufferString(tc.data))
			require.NoError(t, err)
			require.NotNil(t, p)

			if len(tc.tags) > 0 {
				p.WithTags(tc.tags...)
			}

			b, err2 := yaml.Marshal(p)
			require.NoError(t, err2)
			require.NotNil(t, b)
			assert.Equal(t, tc.want, string(b))
		})
	}
}

func TestPolicy_UnmarshalJSON_Fail(t *testing.T) {
	t.Parallel()

	t.Run("invalid json", func(t *testing.T) {
		t.Parallel()

		p := new(Policy)
		err := p.UnmarshalJSON([]byte{1, 2, 3})
		require.Error(t, err)
	})
}

func TestSplitPolicyKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		key   string
		want1 string
		want2 string
	}{
		{name: "empty"},
		{name: "just separator", key: "/"},
		{name: "only id", key: "/x", want2: "x"},
		{name: "simple", key: "x/y", want1: "x", want2: "y"},
		{name: "opa", key: "Opa/X/y", want1: "Opa/X", want2: "y"},
		{name: "cerbos", key: "CERBOS/x/Y", want1: "CERBOS/x", want2: "Y"},
		{name: "cedar", key: "cedar/x/y", want2: "cedar/x/y"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got1, got2 := SplitPolicyKey(tc.key)
			assert.Equal(t, tc.want1, got1)
			assert.Equal(t, tc.want2, got2)
		})
	}
}
