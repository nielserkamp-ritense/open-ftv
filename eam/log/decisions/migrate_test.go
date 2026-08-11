package decisions

import (
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	testCases := []struct {
		name    string
		source  string
		dsn     string
		steps   int
		auto    bool
		wantErr bool
	}{
		{
			// an emptied source is how migrating the ADL is switched off.
			name: "no source",
			dsn:  "not even a dsn",
			auto: true,
		},
		{
			name:   "neither auto nor steps",
			source: "*embed*",
			dsn:    "not even a dsn",
		},
		{
			name:    "embedded scripts with an unusable dsn",
			source:  "*embed*",
			dsn:     "not even a dsn",
			auto:    true,
			wantErr: true,
		},
		{
			name:    "relative file source is made absolute",
			source:  "no/such/scripts",
			dsn:     "not even a dsn",
			steps:   1,
			wantErr: true,
		},
		{
			name:    "file url source is used as given",
			source:  "file:///no/such/scripts",
			dsn:     "not even a dsn",
			steps:   -1,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := Migrate(tc.source, tc.dsn, tc.steps, tc.auto, logger)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
