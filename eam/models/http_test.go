package models

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPRequest(t *testing.T) {
	rt := time.Date(2024, 10, 29, 12, 13, 14, 999000000, time.UTC)
	u, _ := url.Parse("https://google.com/hello/world")
	j, _ := json.Marshal(u)

	testCases := []struct {
		name string
		in   HTTPRequest
		want string
	}{
		{
			name: "empty",
			in:   HTTPRequest{},
			want: `{}`,
		},
		{
			name: "URL",
			in:   HTTPRequest{URL: u},
			want: fmt.Sprintf(`{"url":%s}`, string(j)),
		},
		{
			name: "RequestTime",
			in:   HTTPRequest{RequestTime: &rt},
			want: `{"requestTime":"2024-10-29T12:13:14.999Z"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.in)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.want, string(got))
		})
	}
}
