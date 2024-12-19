package pap

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolicy(t *testing.T) {
	testCases := []struct {
		name    string
		id      string
		source  string
		target  string
		rvvaID  string
		uri     string
		data    string
		wantErr bool
	}{
		{
			name: "only id",
			id:   "x1",
			data: "policy-1",
		},
		{
			name:   "some meta",
			id:     "x2",
			source: "s1",
			target: "t1",
			data:   "policy-2",
		},
		{
			name:   "all meta",
			id:     "x3",
			source: "s3",
			target: "t3",
			rvvaID: "e3",
			uri:    "https://some.site/policies/x3",
			data:   "policy-3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewPolicy(tc.id, tc.source, tc.target, tc.rvvaID, tc.uri, bytes.NewBufferString(tc.data))
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)

				assert.Equal(t, tc.id, got.ID())
				assert.Equal(t, tc.source, got.Source())
				assert.Equal(t, tc.target, got.Target())
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
	path1 := "../../../testdata/unittest/cedar/allow_post.cedar"
	path2 := "../../../testdata/unittest/opa/subsidies.rego"
	path3 := "../../../testdata/unittest/openfga/doelbinding.model"

	f1, err := os.Open(path1)
	require.NoError(t, err)
	f1.Close()

	testCases := []struct {
		name        string
		path        string
		content     io.Reader
		wantErr     bool
		wantID      string
		wantSource  string
		wantTarget  string
		wantRvva    string
		wantURI     string
		wantPath    string
		wantContent string
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
			name:        "with yaml metadata",
			path:        path2,
			content:     bytes.NewBufferString("some data"),
			wantID:      "subsidies",
			wantSource:  "source1",
			wantTarget:  "target1",
			wantRvva:    "rvva1",
			wantURI:     "https://my.site/pol/x1",
			wantPath:    path2,
			wantContent: "some data",
		},
		{
			name:        "with json metadata",
			path:        path3,
			content:     bytes.NewBufferString("some data"),
			wantID:      "doelbinding.model",
			wantSource:  "source2",
			wantTarget:  "target2",
			wantRvva:    "rvva2",
			wantURI:     "https://my.site/openfga/doelbinding.model",
			wantPath:    path3,
			wantContent: "some data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err2 := NewPolicyFromStore(tc.path, tc.content)
			if tc.wantErr {
				require.Error(t, err2)
				require.Nil(t, got)
			} else {
				require.NoError(t, err2)
				require.NotNil(t, got)

				assert.Equal(t, tc.wantID, got.ID())
				assert.Equal(t, tc.wantSource, got.Source())
				assert.Equal(t, tc.wantTarget, got.Target())
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
