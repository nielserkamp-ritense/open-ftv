package config

import "time"

// DecisionLog contains the configuration variables for the Authorization Decision Log.
type DecisionLog struct {
	Type         string        `json:"adlType,omitempty"         yaml:"log.decisions.type,omitempty"                      env:"ADL_TYPE"          flag:"adl-type"          desc:"Type of Authorization Decision Log (supported: postgresql, stdout, stderr, slog, opentelemetry)"`
	Service      string        `json:"adlService,omitempty"      yaml:"log.decisions.service,omitempty"                   env:"ADL_SERVICE_NAME"  flag:"adl-service"       desc:"ADL service name"`
	Timeout      time.Duration `json:"adlTimeout,omitempty"      yaml:"log.decisions.timeout,omitempty"                   env:"ADL_TIMEOUT"       flag:"adl-timeout"       desc:"ADL batch timeout (default 5s)"                                default:"5s"`
	PgURL        string        `json:"-"                         yaml:"log.decisions.postgresql.url,omitempty"            env:"ADL_PG_URL"        flag:"adl-pg-url"        desc:"ADL PostgreSQL server URL"`
	PgMaxLife    time.Duration `json:"adlPgMaxLife,omitempty"    yaml:"log.decisions.postgresql.connection.ttl,omitempty" env:"ADL_PG_MAX_LIFE"   flag:"adl-pg-max-life"   desc:"ADL PostgreSQL inactive connections time-to-live (default 5m)" default:"5m"`
	PgMaxConn    int32         `json:"adlPgMaxConn,omitempty"    yaml:"log.decisions.postgresql.connection.max,omitempty" env:"ADL_PG_MAX_CONN"   flag:"adl-pg-max-conn"   desc:"ADL PostgreSQL maximum connections (default 100)"              default:"100"`
	OtelURL      string        `json:"-"                         yaml:"log.decisions.otel.url,omitempty"                  env:"ADL_OTEL_URL"      flag:"adl-otel-url"      desc:"ADL OpenTelemetry collector URL"`
	OtelInsecure bool          `json:"adlOtelInsecure,omitempty" yaml:"log.decisions.otel.insecure,omitempty"             env:"ADL_OTEL_INSECURE" flag:"adl-otel-insecure" desc:"ADL allow insecure OpenTelemetry connection"`
	SlogMsg      string        `json:"adlSLogMsg,omitempty"      yaml:"log.decisions.slog.message,omitempty"              env:"ADL_SLOG_MESSAGE"  flag:"adl-slog-message"  desc:"ADL log message for slog type"`
	Pretty       bool          `json:"adlLPretty,omitempty"      yaml:"log.decisions.file.prettyPrint,omitempty"          env:"ADL_PRETTY_PRINT"  flag:"adl-pretty-print"  desc:"ADL pretty printing for stdout/stderr type"`

	// Migration of the ADL database. The ADL is not the app's own database, so it has its own
	// migration settings; when Auto and Steps are left unset the app falls back to its general
	// Migration settings. Set the source to an empty value to switch ADL migration off.
	MigrateSource string `json:"adlMigrationSource,omitempty" yaml:"log.decisions.migrate.source,omitempty" env:"ADL_MIGRATE_SOURCE" flag:"adl-migrate-source" desc:"Source location of ADL migration scripts (default embedded scripts)"  default:"*embed*"`
	MigrateSteps  int    `json:"adlMigrationSteps,omitempty"  yaml:"log.decisions.migrate.steps,omitempty"  env:"ADL_MIGRATE_STEPS"  flag:"adl-migrate-steps"  desc:"ADL migration steps to perform (positive is up, negative is down)"`
	MigrateAuto   bool   `json:"adlMigrationAuto,omitempty"   yaml:"log.decisions.migrate.auto,omitempty"   env:"ADL_MIGRATE_AUTO"   flag:"adl-migrate-auto"   desc:"Always migrate the ADL up to the latest level"`
}

// Sanitized returns the configuration variables where all sensitive data has been scrubbed.
func (d *DecisionLog) Sanitized() *DecisionLog {
	sanitized := *d
	sanitized.OtelURL = ""
	sanitized.PgURL = ""
	return &sanitized
}
