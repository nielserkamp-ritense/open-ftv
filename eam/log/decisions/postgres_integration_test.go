//go:build integration

package decisions

import (
	"context"
	"fmt"
	"hash/fnv"
	"log/slog"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/opentelemetry"
)

// Verifies ADR "Safe to re-deliver" (§3.2.4): re-submitting the same record MUST NOT duplicate it. Only
// testable here, not via the PDP's HTTP surface, since that always mints a fresh span_id per request.

const defaultDecisionsTestPostgresURL = "postgres://user:password@localhost:5432/openftv?sslmode=disable"

var nonDatabaseNameChar = regexp.MustCompile(`[^a-z0-9_]+`)

// decisionsTestDatabaseName mirrors the equivalent helper in apps/pdp/server's integration tests, kept as
// a separate copy since that one is private to a different package.
func decisionsTestDatabaseName(name string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	suffix := fmt.Sprintf("_%08x", h.Sum32())

	safe := nonDatabaseNameChar.ReplaceAllString(strings.ToLower(name), "_")

	const maxLengthDatabaseName = 63 // Postgres identifier length limit.
	const prefix = "adl_"

	if maxPrefixLen := maxLengthDatabaseName - len(prefix) - len(suffix); len(safe) > maxPrefixLen {
		safe = safe[:maxPrefixLen]
	}

	return prefix + safe + suffix
}

func withDatabase(dsn, name string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}

	u.Path = "/" + name
	return u.String()
}

// newTestPostgresLogger returns a Logger backed by a fresh, migrated, isolated database, and that
// database's DSN for direct verification queries.
func newTestPostgresLogger(t *testing.T) (*Logger, string) {
	t.Helper()

	baseURL := os.Getenv("MANAGER_TEST_POSTGRES_URL")
	if baseURL == "" {
		baseURL = defaultDecisionsTestPostgresURL
	}

	ctx := context.Background()

	admin, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Skipf("postgres not reachable at %s (start it with `docker compose -f docker/postgres.yaml up -d`): %v", baseURL, err)
	}
	t.Cleanup(func() { _ = admin.Close(ctx) })

	dbName := decisionsTestDatabaseName(t.Name())

	if _, err = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName+" WITH (FORCE)"); err != nil {
		t.Fatalf("failed to drop existing test database %q: %v", dbName, err)
	}
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("failed to create test database %q: %v", dbName, err)
	}

	dsn := withDatabase(baseURL, dbName)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	require.NoError(t, Migrate("*embed*", dsn, 0, true, logger))

	l, err := NewWithPostgreSQL(ctx, "adl-idempotency-test", dsn, nil, time.Minute, 5, opentelemetry.WithBatchTimeout(2*time.Second))
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Shutdown(context.Background()) })

	return l, dsn
}

func countDecisionRows(t *testing.T, dsn, traceID, spanID string) int {
	t.Helper()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	defer conn.Close(ctx)

	var count int
	err = conn.QueryRow(ctx, "SELECT count(*) FROM decision WHERE trace_id = $1 AND span_id = $2", traceID, spanID).Scan(&count)
	require.NoError(t, err)

	return count
}

func TestPgLogger_Decision_IdempotentIngestion(t *testing.T) {
	logger, dsn := newTestPostgresLogger(t)

	const traceID = "9bf92f3577b34da6a3ce929d0e0e473b"
	const spanID = "00f067aa0ba902b7"

	d := &Decision{
		Timestamp:   time.Now().UTC(),
		RequestType: EvaluationEndpoint,
		TraceID:     traceID,
		SpanID:      spanID,
		Request:     map[string]any{"x": 1},
		Response:    map[string]any{"decision": true},
	}

	require.NoError(t, logger.Decision(context.Background(), d))
	require.NoError(t, logger.Decision(context.Background(), d)) // simulated redelivery of the same record

	require.Equal(t, 1, countDecisionRows(t, dsn, traceID, spanID))
}
