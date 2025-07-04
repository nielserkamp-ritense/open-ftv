package config

import (
	"strings"

	authentication2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
)

// Authentication contains the configuration variables for API endpoints to authenticate users and/or external processes.
type Authentication struct {
	Type string `yaml:"authentication.type,omitempty" env:"AUTHENTICATION_TYPE" flag:"authentication-type" desc:"Type of authentication verification (bcrypt)"`
}

// NewAuthenticator instantiates a new authenticator using the given configuration.
func (a *Authentication) NewAuthenticator(controller pdp.Controller) (authentication2.Authenticator, error) {
	opts := []authentication2.Option{
		authentication2.WithContext(controller.Context()),
		authentication2.WithLogger(controller.Logger()),
		authentication2.WithEntities(controller.PIP()),
	}

	// TODO: basic auth & jwt authentication handlers.

	switch strings.ToLower(a.Type) {
	case "bcrypt":
		return authentication2.NewBCrypt(opts...), nil
	case "none", "":
		// !!NOTE!! default is no authentication.
		return authentication2.NewDummy(opts...), nil
	// case "basicauth":
	// 	return authentication2.NewBasicAuth(opts...), nil
	// case "jwt":
	// 	return authentication2.NewJWT(opts...), nil
	default:
		return nil, nil
	}
}
