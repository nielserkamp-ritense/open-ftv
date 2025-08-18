package pap

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestPAP_LoadFiles(t *testing.T) {
	t.Parallel()

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
			path:         "../../testdata/policies/cedar/brp",
			recurse:      true,
			wantPolicies: 3,
		},
		{
			name:         "cedar - recurse",
			path:         "../../testdata/policies/cedar",
			recurse:      true,
			wantPolicies: 5,
		},
		{
			name: "cedar - bad meta",
			path: "../../testdata/unittest/bad/meta",
		},
		{
			name:         "opa - with meta",
			path:         "../../testdata/unittest/opa",
			wantPolicies: 1,
		},
		{
			name: "empty with dot file",
			path: "../../testdata/unittest/empty",
		},
		{
			name: "cedar - no recurse",
			path: "../../testdata/policies/cedar",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			p := New(nil, slog.New(h), WithFileStore(tc.path, tc.recurse))
			require.NotNil(t, p)

			p.LoadFiles()
			assert.GreaterOrEqual(t, h.Count(), tc.wantLog)
		})
	}
}

func TestPAP_LoadFiles_ForceBadPath(t *testing.T) {
	t.Parallel()

	t.Run("load files - force bad path", func(t *testing.T) {
		t.Parallel()

		h := slog2.NewDummyHandler(slog.LevelInfo)

		p := &PAP{policyStore: "/this/is/not/a/directory/at/all/!", logger: slog.New(h)}
		p.LoadFiles()
		assert.GreaterOrEqual(t, h.Count(), 1)
	})
}
