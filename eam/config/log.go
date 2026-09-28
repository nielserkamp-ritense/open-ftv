package config

import "flag"

// Log contains the configuration variables for application level logging.
type Log struct {
	Output string
	Format string
	Level  string
	Source bool
}

func (l *Log) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&l.Output, "log-output", "stdout", "File for writing log data")
	fs.StringVar(&l.Format, "log-format", "json", "Format to use when writing log data (json, text, console)")
	fs.StringVar(&l.Level, "log-level", "info", "Level for writing log data (debug, info, warn, error)")
	fs.BoolVar(&l.Source, "log-source", true, "Include source-location when writing log data")
}
