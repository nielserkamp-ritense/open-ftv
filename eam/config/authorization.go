package config

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/authentication"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
)

// Authorization contains the configuration variables for API endpoints to authorize access for users and/or external processes.
type Authorization struct {
	Authenticate bool `yaml:"authorization.authenticate,omitempty" env:"AUTHORIZATION_AUTHENTICATE" flag:"authorization-authenticate" desc:"Indicates the authorization process must first authenticate the user an/or process"`
}

// NewAuthorizer instantiates a new authorizer using the given configuration.
func (a *Authorization) NewAuthorizer(controller pdp.Controller, authenticator authentication.Authenticator) (authorization.Authorizer, error) {
	opts := []authorization.Option{
		authorization.WithContext(controller.Context()),
		authorization.WithLogger(controller.Logger()),
		authorization.WithPEP(controller.PEP()),
		authorization.WithPDP(controller),
		authorization.WithEntities(controller.PIP()),
	}

	if a.Authenticate {
		if authenticator == nil {
			return nil, fmt.Errorf("failed to initialize authorizer: no authenticator provided")
		}
		opts = append(opts, authorization.WithAuthenticator(authenticator))
	}

	return authorization.New(opts...), nil
}
