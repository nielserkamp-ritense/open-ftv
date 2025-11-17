package models

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequest(t *testing.T) {
	t.Parallel()

	uid, _ := uuid.NewUUID()
	rt := time.Date(2024, 10, 29, 12, 13, 14, 999000000, time.UTC)
	u, _ := url.Parse("https://google.com/hello/world")
	j, _ := json.Marshal(u)

	testCases := []struct {
		name string
		in   *Request
		want string
	}{
		{
			name: "empty",
			in:   &Request{},
			want: `{}`,
		},
		{
			name: "UID",
			in:   &Request{UID: &uid},
			want: fmt.Sprintf(`{"uid":"%s"}`, uid.String()),
		},
		{
			name: "URL",
			in:   &Request{URL: u},
			want: fmt.Sprintf(`{"url":%s}`, string(j)),
		},
		{
			name: "RequestTime",
			in:   &Request{RequestTime: &rt},
			want: `{"requestTime":"2024-10-29T12:13:14.999Z"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.in)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.want, string(got))
		})
	}
}

func TestPARC(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   *PARC
		want string
	}{
		{
			name: "empty",
			in:   &PARC{},
			want: "{}",
		},
		{
			name: "principal without attributes",
			in:   &PARC{Principal: NewEntity("user", "alice", NewAttributeSet())},
			want: `{"principal":{"type":"user","id":"alice","status":"???","attributes":[]}}`,
		},
		{
			name: "principal with attributes",
			in: &PARC{
				Principal: NewEntity("user", "alice",
					NewAttributeSet(
						NewAttribute("x", "y"),
						NewAttribute("z", "w"),
					)),
			},
			want: `{"principal":{"type":"user","id":"alice","status":"???","attributes":[{"key":"x","value":"y","status":"???"},{"key":"z","value":"w","status":"???"}]}}`,
		},
		{
			name: "action without attributes",
			in:   &PARC{Action: NewEntity("method", "POST", NewAttributeSet())},
			want: `{"action":{"type":"method","id":"POST","status":"???","attributes":[]}}`,
		},
		{
			name: "context",
			in: &PARC{Context: NewAttributeSet(
				NewAttribute("x", "y"),
				NewAttribute("y", 123),
				NewAttribute("z", true),
				NewAttribute("z2", 1.345678),
			)},
			want: `{"context":[{"key":"x","value":"y","status":"???"},{"key":"y","value":123,"status":"???"},{"key":"z","value":true,"status":"???"},{"key":"z2","value":1.345678,"status":"???"}]}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := json.Marshal(tc.in)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.want, string(got))
		})
	}
}
