package config

import "flag"

// Log contains the configuration variables for application level logging.
type Log struct {
	Output string `env:"LOG_OUTPUT" flag:"log-output" desc:"File for writing log data"                             default:"stdout"`
	Format string `env:"LOG_FORMAT" flag:"log-format" desc:"Format to use when writing log data (json, text)"      default:"json"`
	Level  string `env:"LOG_LEVEL"  flag:"log-level"  desc:"Level for writing log data (debug, info, warn, error)" default:"info"`
	Source bool   `env:"LOG_SOURCE" flag:"log-source" desc:"Include source-location when writing log data"         default:"true"`
}

func (l *Log) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&l.Output, "log-output", "stdout", "File for writing log data")
	fs.StringVar(&l.Format, "log-format", "json", "Format to use when writing log data (json, text)")
	fs.StringVar(&l.Level, "log-level", "info", "Level for writing log data (debug, info, warn, error)")
	fs.BoolVar(&l.Source, "log-source", true, "Include source-location when writing log data")
}
