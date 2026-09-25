package config

import "flag"

// Authentication contains the configuration variables for API endpoints to authenticate users and/or external processes.
type Authentication struct {
	Type string
}

func (a *Authentication) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&a.Type, "authentication-type", "", "Type of authentication verification (bcrypt)")
}
