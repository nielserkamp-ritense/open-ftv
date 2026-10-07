package server

import (
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/laadpalen-applicatie/config"
)

func TestServeStopsWhenHealthPortTaken(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp4", "127.0.0.1:20013")
	require.NoError(t, err)

	t.Cleanup(func() { _ = ln.Close() })

	cfg := &config.Config{Host: "127.0.0.1", Port: 20012, HealthHost: "127.0.0.1", HealthPort: 20013}

	srv, err := NewService(cfg, slog.New(slog.DiscardHandler), testAssets, nil)
	require.NoError(t, err)

	done := make(chan error)

	go func() {
		done <- srv.Serve()
	}()

	// Serve returns only once the main service has stopped as well.
	select {
	case err := <-done:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the main service kept running without its health service")
	}
}
