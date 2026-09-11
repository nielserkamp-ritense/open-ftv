package server

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// TestNewExternal_MalformedDatabaseURL asserts that a DatabaseURL the connection pool cannot parse
// aborts startup, rather than leaving the app running with a database it can never reach.
func TestNewExternal_MalformedDatabaseURL(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r)
		// Pin the panic to the pool, so a later startup failure cannot satisfy this test.
		require.Contains(t, fmt.Sprint(r), "dsn parsing failed")
	}()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelError))

	cfg := &config.Config{
		PAP:         config2.PAP{Language: "cedar"},
		DatabaseURL: "postgres://user:password@localhost:not-a-port/openftv",
	}

	_ = NewExternal(cfg, logger)

	require.Fail(t, "should never reach here, because we expect NewExternal to panic")
}
