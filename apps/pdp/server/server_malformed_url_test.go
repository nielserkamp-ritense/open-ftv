package server

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config"
)

// TestNewService_MalformedADLDatabaseURL asserts that an ADLDatabaseURL the decision log cannot
// parse aborts startup, rather than leaving the PDP running with a database it can never reach.
func TestNewService_MalformedADLDatabaseURL(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r)
		// Pin the panic to the decision log DSN, so a later startup failure cannot satisfy this test.
		require.Contains(t, fmt.Sprint(r), "failed to initialize authorization decision log")
		require.Contains(t, fmt.Sprint(r), `invalid port ":not-a-port"`)
	}()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelError))

	cfg := &config.Config{
		PAP:            config2.PAP{Language: "cedar"},
		ADLService:     "pdp-test",
		ADLDatabaseURL: "postgres://user:password@localhost:not-a-port/openftv",
	}

	_ = NewService(cfg, logger)

	require.Fail(t, "should never reach here, because we expect NewService to panic")
}
