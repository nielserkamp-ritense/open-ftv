package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAudit(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		created   time.Time
		createdBy string
		updated   time.Time
		updatedBy string
	}{
		{
			name:      "created",
			created:   time.Date(2025, 6, 15, 12, 13, 14, 1500000000, time.UTC),
			createdBy: "ik",
		},
		{
			name:      "updated",
			updated:   time.Date(2025, 6, 15, 12, 13, 14, 1500000000, time.UTC),
			updatedBy: "jij",
		},
		{
			name:      "all",
			created:   time.Date(2025, 6, 15, 12, 13, 14, 1500000000, time.UTC),
			createdBy: "ik",
			updated:   time.Date(2025, 6, 16, 17, 18, 19, 2000000000, time.UTC),
			updatedBy: "jij",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := &Audit{created: tc.created, createdBy: tc.createdBy, updated: tc.updated, updatedBy: tc.updatedBy}
			assert.Equal(t, tc.created, a.Created())
			assert.Equal(t, tc.createdBy, a.CreatedBy())
			assert.Equal(t, tc.updated, a.Updated())
			assert.Equal(t, tc.updatedBy, a.UpdatedBy())
		})
	}
}
