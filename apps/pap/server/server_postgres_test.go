//go:build integration

package server

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

const defaultTestPostgresURL = "postgres://user:password@localhost:5432/openftv?sslmode=disable"

// newTestDatabaseURL creates an empty database for the calling test and returns its URL. The PAP
// migrates it to the latest level itself, at startup.
//
// It skips the test when no Postgres is reachable, so the suite still runs without one; point
// PAP_TEST_POSTGRES_URL at a server whose credentials differ from the compose file's.
func newTestDatabaseURL(t *testing.T) string {
	t.Helper()

	baseURL := os.Getenv("PAP_TEST_POSTGRES_URL")
	if baseURL == "" {
		baseURL = defaultTestPostgresURL
	}

	ctx := context.Background()

	admin, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Skipf("postgres not reachable at %s (start it with `docker compose -f docker/compose-integration-tests.yaml up -d`): %v", baseURL, err)
	}
	t.Cleanup(func() { _ = admin.Close(ctx) })

	dbName := composeDatabaseNameFromTestName(t, t.Name())

	if _, err = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName+" WITH (FORCE)"); err != nil {
		t.Fatalf("failed to drop existing test database %q: %v", dbName, err)
	}

	if _, err = admin.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("failed to create test database %q: %v", dbName, err)
	}

	return withDatabase(baseURL, dbName)
}

func composeDatabaseNameFromTestName(t *testing.T, name string) string {
	t.Helper()

	name = strings.ToLower(name)
	name = strings.Replace(name, "test", "", 1) // strip 'test'-prefix
	name = strings.ReplaceAll(name, "/", "_")   // subtest separators

	db := "pap_" + name

	// via https://stackoverflow.com/a/27865772/363448
	const maxLengthDatabaseName = 63 // Postgres identifier length limit.
	if len(db) > maxLengthDatabaseName {
		t.Fatalf("database name too long (%d > %d characters): %q", len(db), maxLengthDatabaseName, db)
	}

	return db
}

func withDatabase(dsn, name string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}

	u.Path = "/" + name

	return u.String()
}
