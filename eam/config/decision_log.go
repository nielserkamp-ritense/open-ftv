package config

import "time"

// DecisionLog contains the configuration variables for the Authorization Decision Log.
type DecisionLog struct {
	Type         string        `yaml:"log.decisions.type,omitempty"                      env:"ADL_TYPE"          flag:"adl-type"          desc:"Type of Authorization Decision Log (supported: postgresql, stdout, stderr, slog, opentelemetry)"`
	Service      string        `yaml:"log.decisions.service,omitempty"                   env:"ADL_SERVICE_NAME"  flag:"adl-service"       desc:"ADL service name"`
	Timeout      time.Duration `yaml:"log.decisions.timeout,omitempty"                   env:"ADL_TIMEOUT"       flag:"adl-timeout"       desc:"ADL batch timeout (default 5s)" default:"5s"`
	PgURL        string        `yaml:"log.decisions.postgresql.url,omitempty"            env:"ADL_PG_URL"        flag:"adl-pg-url"        desc:"ADL PostgreSQL server URL"`
	PgMaxLife    time.Duration `yaml:"log.decisions.postgresql.connection.ttl,omitempty" env:"ADL_PG_MAX_LIFE"   flag:"adl-pg-max-life"   desc:"ADL PostgreSQL inactive connections time-to-live (default 5m)" default:"5m"`
	PgMaxConn    int32         `yaml:"log.decisions.postgresql.connection.max,omitempty" env:"ADL_PG_MAX_CONN"   flag:"adl-pg-max-conn"   desc:"ADL PostgreSQL maximum connections (default 100)"              default:"100"`
	OtelURL      string        `yaml:"log.decisions.otel.url,omitempty"                  env:"ADL_OTEL_URL"      flag:"adl-otel-url"      desc:"ADL OpenTelemetry collector URL"`
	OtelInsecure bool          `yaml:"log.decisions.otel.insecure,omitempty"             env:"ADL_OTEL_INSECURE" flag:"adl-otel-insecure" desc:"ADL allow insecure OpenTelemetry connection"`
	SlogMsg      string        `yaml:"log.decisions.slog.message,omitempty"              env:"ADL_SLOG_MESSAGE"  flag:"adl-slog-message"  desc:"ADL log message for slog type"`
	Pretty       bool          `yaml:"log.decisions.file.prettyPrint,omitempty"          env:"ADL_PRETTY_PRINT"  flag:"adl-pretty-print"  desc:"ADL pretty printing for stdout/stderr type"`
}

// Sanitized returns the configuration variables where all sensitive data has been scrubbed.
func (d *DecisionLog) Sanitized() *DecisionLog {
	sanitized := *d
	sanitized.OtelURL = ""
	sanitized.PgURL = ""
	return &sanitized
}
