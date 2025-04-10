package config

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
)

// Authentication contains the configuration variables for API endpoints to authenticate users and/or external processes.
type Authentication struct {
	Type string `yaml:"authentication.type,omitempty" env:"AUTHENTICATION_TYPE" flag:"authentication-type" desc:"Type of authentication verification (bcrypt)"`
}

// NewAuthenticator instantiates a new authenticator using the given configuration.
func (a *Authentication) NewAuthenticator(controller pdp.Controller) (authentication.Authenticator, error) {
	opts := []authentication.Option{
		authentication.WithContext(controller.Context()),
		authentication.WithLogger(controller.Logger()),
		authentication.WithEntities(controller.PIP()),
	}

	switch strings.ToLower(a.Type) {
	case "bcrypt":
		return authentication.NewBCrypt(opts...), nil
	default:
		return nil, nil
	}
}
