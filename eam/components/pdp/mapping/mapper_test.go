package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapper_Configure(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		opts           []Option
		wantHeaderKeys []string
	}{
		{name: "none"},
		{name: "header key", opts: []Option{WithHeaderKeys("abc")}, wantHeaderKeys: []string{"abc"}},
		{name: "header keys", opts: []Option{WithHeaderKeys("abc", "def", "ghi")}, wantHeaderKeys: []string{"abc", "def", "ghi"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := &base{}
			m.configure(tc.opts)
			assert.EqualValues(t, tc.wantHeaderKeys, m.headerKeys)
		})
	}
}

func TestMappingsFromConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		cfg       string
		wantCount int
	}{
		{name: "empty"},
		{name: "rvva", cfg: "xyz , RvVAtoPrincipal ", wantCount: 1},
		{name: "doelbinding", cfg: " goed-doel , DoelBindingToPrincipal", wantCount: 1},
		{name: "body", cfg: "xyz , BodyToContext ", wantCount: 1},
		{name: "all", cfg: "  xyz, DOELBINDINGtoPrincipal, BODYtoContext, abc,RVVAtoPrincipal", wantCount: 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := MappingsFromConfig(tc.cfg)
			assert.Equal(t, tc.wantCount, len(got))
		})
	}
}
