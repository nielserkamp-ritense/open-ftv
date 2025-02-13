package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusCode_Match(t *testing.T) {
	testCases := []struct {
		name   string
		code   string
		status int
		want   bool
	}{
		{name: "zero", code: "2xx"},
		{name: "10", code: "2xx", status: 10},
		{name: "199", code: "2xx", status: 199},
		{name: "200", code: "2xx", status: 200, want: true},
		{name: "200 exact", code: "200", status: 200, want: true},
		{name: "201", code: "2xx", status: 201, want: true},
		{name: "201 exact", code: "200", status: 201},
		{name: "299", code: "2xx", status: 299, want: true},
		{name: "300", code: "2xx", status: 300},
		{name: "1200", code: "2xx", status: 1200},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &StatusCode{Code: tc.code}
			got := s.Match(tc.status)
			assert.Equal(t, tc.want, got)
		})
	}
}
