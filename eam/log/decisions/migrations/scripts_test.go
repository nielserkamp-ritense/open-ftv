package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		file    string
		wantErr bool
	}{
		{name: "bad file", file: "xyz.sql", wantErr: true},
		{name: "good file", file: "postgresql/00001_initial.down.sql"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := PgDecisionLog.ReadFile(tc.file)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, b)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, b)
			}
		})
	}
}
