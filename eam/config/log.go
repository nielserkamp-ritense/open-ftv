package config

import (
	"log/slog"

	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

// Log contains the configuration variables for application level logging.
type Log struct {
	Output string `yaml:"log.output,omitempty" env:"LOG_OUTPUT" flag:"log-output" default:"stdout" desc:"File for writing log data"`
	Format string `yaml:"log.format,omitempty" env:"LOG_FORMAT" flag:"log-format" default:"json" desc:"Format to use when writing log data (json, text)"`
	Level  string `yaml:"log.level,omitempty" env:"LOG_LEVEL" flag:"log-level" default:"info" desc:"Level for writing log data (debug, info, warn, error)"`
	Source bool   `yaml:"log.source,omitempty" env:"LOG_SOURCE" flag:"log-source" default:"true" desc:"Include source-location when writing log data"`
}

// MakeLogger returns a new logger based on the variables in the struct.
func (l *Log) MakeLogger() (*slog.Logger, error) {
	return slog2.Init(l.Output, l.Format, l.Level, l.Source)
}
