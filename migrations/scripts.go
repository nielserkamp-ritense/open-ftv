// Package migrations contains an embeddable filesystem for all database migration scripts.
package migrations

import "embed"

// PostgreSQL contains all the migration scripts for PostgreSQL.
//
//go:embed "postgresql/*.sql"
var PostgreSQL embed.FS
