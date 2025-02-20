package pap

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/storage/valkeyrie/memory"
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

			s := memory.New()

			c := &pap{
				logger:  slog.New(h),
				persist: NewStore(nil, s, ""),
			}

			c.LoadFromStore(tc.path, tc.recurse)

			assert.GreaterOrEqual(t, tc.wantLog, h.Count())
		})
	}
}
