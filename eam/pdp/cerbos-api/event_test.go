package cerbos_api

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestController_Handle_NoAdminAPI(t *testing.T) {
	t.Run("handle - no admin PI", func(t *testing.T) {
		h := slog2.NewDummyHandler(slog.LevelDebug)
		logger := slog.New(h)

		c := NewController(
			Config{Addr2: "\000\001"},
			pdp.WithLogger(logger),
		)

		c2, ok := c.(*controller)
		require.True(t, ok)
		require.NotNil(t, c2)

		h.Clear()
		c2.Handle(models.PolicyAdded, "x")
		assert.Equal(t, 1, h.Count())
	})
}

func TestController_Handle_BadLanguage(t *testing.T) {
	addr := getAddress()

	testCases := []struct {
		name  string
		event models.EventType
		key   string
	}{
		{
			name:  "add",
			event: models.PolicyAdded,
			key:   "cedar/xyz",
		},
		{
			name:  "replace",
			event: models.PolicyReplaced,
			key:   "cedar/xyz",
		},
		{
			name:  "removed",
			event: models.PolicyRemoved,
			key:   "cedar/xyz",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			c := NewController(
				Config{Addr1: addr, Addr2: addr, User: adminUser, Pswd: adminPswd},
				pdp.WithLogger(logger),
			)

			c2, ok := c.(*controller)
			require.True(t, ok)
			require.NotNil(t, c2)

			h.Clear()
			c2.Handle(tc.event, tc.key)
			assert.Equal(t, 0, h.Count())
		})
	}
}
