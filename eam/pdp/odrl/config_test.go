package odrl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriberList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "single", in: "https://a.example/v1/odrl/events", want: []string{"https://a.example/v1/odrl/events"}},
		{name: "multiple with spaces", in: " https://a.example/e , https://b.example/e ,", want: []string{"https://a.example/e", "https://b.example/e"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := &Config{Subscribers: tc.in}
			assert.Equal(t, tc.want, c.SubscriberList())
		})
	}
}
