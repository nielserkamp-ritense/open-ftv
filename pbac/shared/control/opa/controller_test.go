package opa

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestNewController(t *testing.T) {
	testCases := []struct {
		name     string
		store1   string
		recurse1 bool
		store2   string
		recurse2 bool
		wantLog  int
	}{
		{
			name:    "no stores",
			wantLog: 2,
		},
		{
			name:    "pip store - no recurse",
			store1:  "../../../../testdata/pip",
			wantLog: 2,
		},
		{
			name:     "pip store - recurse",
			store1:   "../../../../testdata/pip",
			recurse1: true,
			wantLog:  3,
		},
		{
			name:    "pap store - no recurse",
			store2:  "../../../../testdata/policies/opa",
			wantLog: 2,
		},
		{
			name:     "pap store - recurse",
			store2:   "../../../../testdata/policies/opa",
			recurse2: true,
			wantLog:  4,
		},
		{
			name:     "pip & pap store - recurse",
			store1:   "../../../../testdata/pip",
			recurse1: true,
			store2:   "../../../../testdata/policies/opa",
			recurse2: true,
			wantLog:  5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			p := pip.New(nil, tc.store1, tc.recurse1, logger, standards.NewAttributeSet, standards.NewEntitySet)
			require.NotNil(t, p)

			h.Clear()

			c := NewController(p, tc.store2, tc.recurse2, logger, nil)
			require.NotNil(t, c)

			assert.Equal(t, tc.wantLog, h.Count())
		})
	}
}
