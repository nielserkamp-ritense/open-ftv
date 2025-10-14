package config

import (
	"log/slog"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

// Log contains the configuration variables for application level logging.
type Log struct {
	Output string `json:"logOutput,omitempty" yaml:"log.output,omitempty" env:"LOG_OUTPUT" flag:"log-output" desc:"File for writing log data"                             default:"stdout"`
	Format string `json:"logFormat,omitempty" yaml:"log.format,omitempty" env:"LOG_FORMAT" flag:"log-format" desc:"Format to use when writing log data (json, text)"      default:"json"`
	Level  string `json:"logLevel,omitempty"  yaml:"log.level,omitempty"  env:"LOG_LEVEL"  flag:"log-level"  desc:"Level for writing log data (debug, info, warn, error)" default:"info"`
	Source bool   `json:"logSource,omitempty" yaml:"log.source,omitempty" env:"LOG_SOURCE" flag:"log-source" desc:"Include source-location when writing log data"         default:"true"`
}

// MakeLogger returns a new logger based on the variables in the struct.
func (l *Log) MakeLogger() (*slog.Logger, error) {
	return slog2.Init(l.Output, l.Format, l.Level, l.Source)
}
