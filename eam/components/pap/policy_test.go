package pap

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/policies"
)

func TestNewPolicy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		p       *policies.Policy
		data    string
		wantErr bool
	}{
		{
			name: "only id",
			p:    &policies.Policy{Id: "x1"},
			data: "policy-1",
		},
		{
			name: "some meta",
			p:    &policies.Policy{Id: "x2"},
			data: "policy-2",
		},
		{
			name: "all meta",
			p:    &policies.Policy{Id: "x3", Language: "opa", RvvaId: "e3", Url: "https://some.site/policies/x3"},
			data: "policy-3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewPolicy(tc.p, bytes.NewBufferString(tc.data))
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

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

			got, err := NewPolicyFromData(tc.id, tc.language, tc.rvvaID, tc.uri, bytes.NewBufferString(tc.data))
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

	path1 := "../../../testdata/unittest/cedar/allow_post.cedar"
	path2 := "../../../testdata/unittest/opa/subsidies.rego"
	path3 := "../../../testdata/unittest/openfga/doelbinding.model"

	f1, err := os.Open(path1)
	require.NoError(t, err)
	f1.Close()

	testCases := []struct {
		name         string
		path         string
		content      io.Reader
		wantErr      bool
		wantID       string
		wantLanguage string
		wantRvva     string
		wantURI      string
		wantPath     string
		wantContent  string
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
			wantPath:    path1,
			wantContent: "some data",
		},
		{
			name:         "with yaml metadata",
			path:         path2,
			content:      bytes.NewBufferString("some data"),
			wantID:       "subsidies",
			wantLanguage: "rego",
			wantRvva:     "rvva1",
			wantURI:      "https://my.site/pol/x1",
			wantPath:     path2,
			wantContent:  "some data",
		},
		{
			name:         "with json metadata",
			path:         path3,
			content:      bytes.NewBufferString("some data"),
			wantID:       "doelbinding.model",
			wantLanguage: "openfga",
			wantRvva:     "rvva2",
			wantURI:      "https://my.site/openfga/doelbinding.model",
			wantPath:     path3,
			wantContent:  "some data",
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
