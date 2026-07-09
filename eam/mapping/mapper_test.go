package mapping

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// resetRegistry clears the global registry and restores it afterwards, so that
// tests can register their own mappers without leaking state between tests.
func resetRegistry(t *testing.T) {
	t.Helper()

	saved := registry
	registry = map[string]Mapper{}
	t.Cleanup(func() { registry = saved })
}

func passthrough(parc *models.PARC, _ ...Option) *models.PARC { return parc }

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

func TestRegister(t *testing.T) {
	resetRegistry(t)

	Register("Alpha", passthrough)
	Register(" beta ", passthrough)

	// names are stored case-insensitively and trimmed.
	assert.Contains(t, registry, "alpha")
	assert.Contains(t, registry, "beta")

	// duplicate registration panics.
	assert.Panics(t, func() { Register("ALPHA", passthrough) })

	// empty name and nil mapper panic.
	assert.Panics(t, func() { Register("", passthrough) })
	assert.Panics(t, func() { Register("gamma", nil) })
}

func TestResolve(t *testing.T) {
	resetRegistry(t)

	Register("alpha", passthrough)
	Register("beta", passthrough)

	testCases := []struct {
		name        string
		cfg         string
		wantCount   int
		wantUnknown []string
	}{
		{name: "empty"},
		{name: "known", cfg: " ALPHA ", wantCount: 1},
		{name: "unknown only", cfg: "xyz, foo", wantUnknown: []string{"xyz", "foo"}},
		{name: "mixed", cfg: "  xyz, Alpha, BETA, abc", wantCount: 2, wantUnknown: []string{"xyz", "abc"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mappers, unknown := Resolve(tc.cfg)
			assert.Len(t, mappers, tc.wantCount)
			assert.Equal(t, tc.wantUnknown, unknown)
		})
	}
}

func TestMappingsFromConfig(t *testing.T) {
	resetRegistry(t)

	Register("alpha", passthrough)

	got := MappingsFromConfig("xyz, Alpha")
	require.Len(t, got, 1)
}
