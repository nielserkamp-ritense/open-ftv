package bundles

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		in       Status
		wantName string
	}{
		{name: "zero", in: 0, wantName: "???"},
		{name: "unknown", in: 99, wantName: "???"},
		{name: "creating", in: Creating, wantName: "Creating new deployment"},
		{name: "gathering", in: Gathering, wantName: "Gathering policies & data"},
		{name: "versioning", in: Versioning, wantName: "Calculating new deployment version"},
		{name: "merging", in: Merging, wantName: "Merging new deployment version in Git"},
		{name: "bundling", in: Bundling, wantName: "Bundling policies & data"},
		{name: "sending", in: Sending, wantName: "Sending bundles"},
		{name: "completed", in: Completed, wantName: "Finished successfully"},
		{name: "min", in: StatusMIN, wantName: "Creating new deployment"},
		{name: "max", in: StatusMAX, wantName: "Finished successfully"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := tc.in.String()
			assert.Equal(t, tc.wantName, got)
		})
	}
}
