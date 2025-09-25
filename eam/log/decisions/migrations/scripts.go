// Package migrations contains an embeddable filesystem for migration scripts for OpenTelemetry.
package migrations

import "embed"

// PgDecisionLog contains the migration scripts for PostgreSQL for the Authorization Decision Log.
//
//go:embed "postgresql/*.sql"
var PgDecisionLog embed.FS
