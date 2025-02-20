package pap

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestCache_LoadFromStore(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		recurse      bool
		wantLog      int
		wantPolicies int
	}{
		{
			name:    "empty",
			recurse: true,
			wantLog: 0,
		},
		{
			name:    "invalid",
			path:    "/this/is/not/a/directory",
			recurse: true,
			wantLog: 1,
		},
		{
			name:         "cedar/brp",
			path:         "../../../testdata/policies/cedar/brp",
			recurse:      true,
			wantPolicies: 3,
		},
		{
			name:         "cedar - recurse",
			path:         "../../../testdata/policies/cedar",
			recurse:      true,
			wantPolicies: 5,
		},
		{
			name: "cedar - no recurse",
			path: "../../../testdata/policies/cedar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelInfo)

			p := New(nil, slog.New(h), WithFileStore(tc.path, tc.recurse))
			require.NotNil(t, p)

			p.LoadFiles()
			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}
