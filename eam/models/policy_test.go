package models

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

func TestNewPolicy(t *testing.T) {
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
			p:       &policies.Policy{Id: "x2"},
			data:    "policy-2",
			wantKey: "/x2",
		},
		{
			name:    "all meta",
			p:       &policies.Policy{Id: "x3", Language: "opa", RvvaId: "e3", Url: "https://some.site/policies/x3"},
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

			got, err := NewPolicy(tc.p, d)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.Equal(t, tc.wantKey, got.Key())
				assert.Equal(t, tc.p.Id, got.ID())
				assert.Equal(t, tc.p.Language, got.Language())
				assert.Equal(t, tc.p.RvvaId, got.RvvaID())
				assert.Equal(t, tc.p.Url, got.URI())

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
		wantDescription string
		wantTags        []string
		wantLanguage    string
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
			name:         "with json metadata",
			path:         path3,
			content:      bytes.NewBufferString("some data"),
			wantID:       "doelbinding.model",
			wantTags:     []string{},
			wantLanguage: "openfga",
			wantRvva:     "rvva2",
			wantURI:      "https://my.site/openfga/doelbinding.model",
			wantPath:     path3,
			wantContent:  "some data",
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
				assert.Equal(t, tc.wantDescription, got.Description())
				assert.EqualValues(t, tc.wantTags, got.Tags())
				assert.Equal(t, tc.wantLanguage, got.Language())
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

func TestPolicy_AddTags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		id       string
		language string
		tags     []string
	}{
		{name: "single", id: "p1", language: "cedar", tags: []string{"x"}},
		{name: "few", id: "p2", language: "rego", tags: []string{"x", "y", "z"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewPolicyFromData(tc.id, tc.language, "", "", bytes.NewBufferString("yo"))
			require.NoError(t, err)
			require.NotNil(t, got)

			got.AddTags(tc.tags...)
			assert.EqualValues(t, tc.tags, got.Tags())

			for i := range tc.tags {
				assert.True(t, got.HasTag(tc.tags[i]))
			}

			assert.False(t, got.HasTag("qqq"))
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
			p:    &policies.Policy{Id: "x3", Language: "opa", RvvaId: "e3", Url: "https://some.site/policies/x3"},
			tags: []string{"x", "y", "z"},
			data: "policy-3",
			want: `{"language":"opa","id":"x3","tags":["x","y","z"],"rvvaID":"e3","uri":"https://some.site/policies/x3","content":"cG9saWN5LTM="}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := NewPolicy(tc.p, bytes.NewBufferString(tc.data))
			require.NoError(t, err)
			require.NotNil(t, p)

			if len(tc.tags) > 0 {
				p.AddTags(tc.tags...)
			}

			b, err2 := json.Marshal(p)
			require.NoError(t, err2)
			require.NotNil(t, b)
			assert.Equal(t, tc.want, string(b))

			p2 := new(Policy)
			err = json.Unmarshal([]byte(b), p2)
			require.NoError(t, err)
			assert.EqualValues(t, p, p2)
		})
	}
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
