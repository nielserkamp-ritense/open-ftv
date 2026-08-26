package adl

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	m1 := map[string]any{"x": "hello world", "y": 123.456}
	m2 := map[string]any{"y": "hello jupiter", "z": true}
	m3 := map[string]any{"engine": "Cerbos", "version": "v1.6.3"}
	m4 := map[string]any{"service": "vlierdam"}

	testCases := []struct {
		name         string
		opts         []Option
		wantVersion  uint64
		wantInfo     any
		wantEngine   any
		wantResource any
	}{
		{name: "none", wantVersion: 0},
		{name: "bundle version", opts: []Option{WithBundleVersion(123)}, wantVersion: 123},
		{name: "information", opts: []Option{WithInformation(m1)}, wantInfo: m1},
		{name: "engine", opts: []Option{WithEngine(m3)}, wantEngine: m3},
		{name: "resource", opts: []Option{WithResource(m4)}, wantResource: m4},
		{
			name:         "all",
			opts:         []Option{WithEngine(m3), WithInformation(m2), WithResource(m4), WithBundleVersion(99)},
			wantVersion:  99,
			wantInfo:     m2,
			wantEngine:   m3,
			wantResource: m4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			logger, err := decisions.New(ctx, "test", opentelemetry.WithFile(os.Stdout, true))
			require.NoError(t, err)
			require.NotNil(t, logger)
			got := New(logger, tc.opts...)
			require.NotNil(t, got)
			assert.Equal(t, tc.wantVersion, got.bundleVersion)
			assert.EqualValues(t, tc.wantInfo, got.information)
			assert.EqualValues(t, tc.wantEngine, got.engine)
			assert.EqualValues(t, tc.wantResource, got.resource)
		})
	}
}
