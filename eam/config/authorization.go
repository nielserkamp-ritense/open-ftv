package config

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	authorization2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
)

// Authorization contains the configuration variables for API endpoints to authorize access for users and/or external processes.
type Authorization struct {
	Authenticate bool `yaml:"authorization.authenticate,omitempty" env:"AUTHORIZATION_AUTHENTICATE" flag:"authorization-authenticate" desc:"Indicates the authorization process must first authenticate the user an/or process"`
}

// NewAuthorizer instantiates a new authorizer using the given configuration.
func (a *Authorization) NewAuthorizer(controller pdp.Controller, authenticator authentication.Authenticator) (authorization2.Authorizer, error) {
	opts := []authorization2.Option{
		authorization2.WithContext(controller.Context()),
		authorization2.WithLogger(controller.Logger()),
		authorization2.WithPEP(controller.PEP()),
		authorization2.WithPDP(controller),
		authorization2.WithEntities(controller.PIP()),
	}

	if a.Authenticate {
		if authenticator == nil {
			return nil, fmt.Errorf("failed to initialize authorizer: no authenticator provided")
		}
		opts = append(opts, authorization2.WithAuthenticator(authenticator))
	}

	return authorization2.New(opts...), nil
}
