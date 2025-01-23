package rdf

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromString(t *testing.T) {
	html := "<html><head><title>Hello World</title></head><body>nothing to see here</body></html>"
	xml := "<xml><title>Hello World</title></xml>"
	j1 := `{"title":"Hello World"}`
	j2 := map[string]any{"title": "Hello World"}

	testCases := []struct {
		name    string
		data    string
		t       string
		want    any
		wantErr bool
	}{
		{name: "no mime-type", data: "abc", wantErr: true},
		{name: "bad mime-type", data: "abc", t: "abc", wantErr: true},
		{name: "string", data: "abc", t: "http://www.w3.org/2001/XMLSchema#string", want: "abc"},
		{name: "html", data: html, t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#HTML", want: html},
		{name: "xml", data: xml, t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral", want: xml},
		{name: "json", data: j1, t: "http://www.w3.org/1999/02/22-rdf-syntax-ns#JSON", want: j2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := FromString(tc.data, tc.t)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}
