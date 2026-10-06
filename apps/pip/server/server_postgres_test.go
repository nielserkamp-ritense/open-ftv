//go:build integration

package server

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/migrations"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/migrate"
)

const defaultTestPostgresURL = "postgres://user:password@localhost:5432/openftv?sslmode=disable"

// newTestDatabaseURL creates an empty database for the calling test, migrates it to the latest
// level with the embedded migration scripts and returns its URL. The PIP does not migrate the
// database itself (the Manager or PAP does), so the test has to do it.
func newTestDatabaseURL(t *testing.T) string {
	t.Helper()

	baseURL := os.Getenv("PIP_TEST_POSTGRES_URL")
	if baseURL == "" {
		baseURL = defaultTestPostgresURL
	}

	ctx := context.Background()

	admin, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Fatalf("postgres not reachable: %v", err)
	}
	t.Cleanup(func() { _ = admin.Close(ctx) })

	dbName := composeDatabaseNameFromTestName(t, t.Name())

	if _, err = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName+" WITH (FORCE)"); err != nil {
		t.Fatalf("failed to drop existing test database %q: %v", dbName, err)
	}

	if _, err = admin.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("failed to create test database %q: %v", dbName, err)
	}

	dsn := withDatabase(baseURL, dbName)

	if err = migrate.PostgresEmbedded(migrations.PostgreSQL, dsn, 0, slog.Default()); err != nil {
		t.Fatalf("failed to migrate test database %q: %v", dbName, err)
	}

	return dsn
}

func composeDatabaseNameFromTestName(t *testing.T, name string) string {
	t.Helper()

	name = strings.ToLower(name)
	name = strings.Replace(name, "test", "", 1) // strip 'test'-prefix
	name = strings.ReplaceAll(name, "/", "_")   // subtest separators

	db := "pip_" + name

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
