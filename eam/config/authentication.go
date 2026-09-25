package config

// Authentication contains the configuration variables for API endpoints to authenticate users and/or external processes.
type Authentication struct {
	Type string `env:"AUTHENTICATION_TYPE" flag:"authentication-type" desc:"Type of authentication verification (bcrypt)"`
}
