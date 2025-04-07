package config

import "time"

// OpenTel contains the configuration variables for an OpenTelemetry sink.
type OpenTel struct {
	URL     string        `yaml:"otel.url,omitempty" env:"OTEL_URL" flag:"otel-url" desc:"OpenTelemetry target URL"`
	Service string        `yaml:"otel.service,omitempty" env:"OTEL_SERVICE_NAME" flag:"otel-service" desc:"OpenTelemetry service name"`
	Timeout time.Duration `yaml:"otel.timeout,omitempty" env:"OTEL_TIMEOUT" flag:"otel-timeout" default:"5s" desc:"OpenTelemetry batch timeout for LDV"`
	Pretty  bool          `yaml:"otel.prettyPrint,omitempty" env:"OTEL_PRETTY_PRINT" flag:"otel-pretty-print" desc:"Facilitate pretty printing of OpenTelemetry log events"`
}
