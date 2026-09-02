//go:build integration

package server

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config"
	config2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config"
)

const defaultTestPostgresURL = "postgres://user:password@localhost:5432/openftv?sslmode=disable"

func newManagerConfigWithPostgres(t *testing.T) *config.Config {
	t.Helper()

	baseURL := os.Getenv("MANAGER_TEST_POSTGRES_URL")
	if baseURL == "" {
		baseURL = defaultTestPostgresURL
	}

	ctx := context.Background()

	admin, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Skipf("postgres not reachable at %s (start it with `make postgres-up`): %v", baseURL, err)
	}
	t.Cleanup(func() { _ = admin.Close(ctx) })

	dbName := composeDatabaseNameFromTestName(t, t.Name())

	if _, err = admin.Exec(ctx, "DROP DATABASE IF EXISTS "+dbName+" WITH (FORCE)"); err != nil {
		t.Fatalf("failed to drop existing test database %q: %v", dbName, err)
	}

	if _, err = admin.Exec(ctx, "CREATE DATABASE "+dbName); err != nil {
		t.Fatalf("failed to create test database %q: %v", dbName, err)
	}

	return &config.Config{
		PAP: config2.PAP{
			Language: "CEDAR",
		},
		Persist: config2.Persist{
			Type:      "postgres",
			PgURL:     withDatabase(baseURL, dbName),
			PgMaxLife: time.Minute,
			PgMaxConn: 10,
		},
		Migration: config2.Migration{
			Source: "*embed*",
			Auto:   true,
		},
	}
}

func composeDatabaseNameFromTestName(t *testing.T, name string) string {
	t.Helper()

	name = strings.ToLower(name)
	name = strings.Replace(name, "test", "", 1) // strip 'test'-prefix
	name = strings.ReplaceAll(name, "/", "_")   // subtest separators

	db := "mgr_" + name

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
