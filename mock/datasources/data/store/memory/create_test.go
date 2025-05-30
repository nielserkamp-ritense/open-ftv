package memory

import (
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store/memory/models"
)

func TestStorage_CreateRecord(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		rec     map[string]any
		t       string
		wantErr bool
	}{
		{
			name:    "bad table",
			rec:     map[string]any{},
			t:       "xyz",
			wantErr: true,
		},
		{
			name: "all good",
			rec:  map[string]any{"kenteken": "AA-99-BB"},
			t:    "rdw.kenteken",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := New(mockFDS)
			require.NotNil(t, s)

			rec := &models.Row{Data: tc.rec}

			err := s.CreateRecord(tc.t, rec)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
