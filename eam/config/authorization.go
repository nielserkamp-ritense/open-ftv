package config

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	authorization2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
)

// Authorization contains the configuration variables for API endpoints to authorize access for users and/or external processes.
type Authorization struct {
	Authenticate bool `json:"forceAuthentication" yaml:"authorization.authenticate,omitempty" env:"AUTHORIZATION_AUTHENTICATE" flag:"authorization-authenticate" desc:"Indicates the authorization process must first authenticate the user an/or process"`
	// FailClosedOnEmpty makes an empty or unreadable policy store DENY all requests
	// instead of allowing them. Default false preserves the legacy fail-open behavior
	// (no policies => NoAuth => allow all). The management plane (manager) enables this
	// so a fresh/wiped/unreadable store locks down rather than exposing the API.
	FailClosedOnEmpty bool `json:"failClosedOnEmpty,omitempty" yaml:"authorization.fail_closed_on_empty,omitempty" env:"AUTHORIZATION_FAIL_CLOSED_ON_EMPTY" flag:"authorization-fail-closed-on-empty" desc:"Deny all requests when the policy store is empty or unreadable, instead of allowing them (fail closed)"`
}

// NewAuthorizer instantiates a new authorizer using the given configuration.
// NewAuthorizer builds an authorizer. The extra options let a caller add concerns this shared
// constructor should not know about, such as the manager passing its principal store.
func (a *Authorization) NewAuthorizer(controller pdp.Controller, authenticator authentication.Authenticator, extra ...authorization2.Option) (authorization2.Authorizer, error) {
	opts := []authorization2.Option{
		authorization2.WithContext(controller.GetContext()),
		authorization2.WithLogger(controller.GetLogger()),
		authorization2.WithPEP(controller.GetPEP()),
		authorization2.WithPDP(controller),
		authorization2.WithEntityGetter(controller.GetPIP().GetEntity),
	}

	opts = append(opts, extra...)

	// Decide whether the policy store is empty/unreadable (or absent entirely). A nil PAP
	// is treated as empty rather than dereferenced, so a misconfigured controller follows
	// the same fail-open/fail-closed path instead of panicking.
	empty := true
	if p := controller.GetPAP(); p != nil {
		list, err := p.List(p.Language().Language())
		empty = err != nil || len(list) == 0
		if err != nil {
			controller.GetLogger().Warn("authorization: failed to read policy store", "err", err)
		}
	}

	if empty {
		if a.FailClosedOnEmpty {
			// Fail closed: do NOT install NoAuth. The PDP evaluates against an empty
			// policy set and denies (default-deny), so a missing/unreadable store locks
			// the API down instead of opening it.
			controller.GetLogger().Warn("authorization: policy store empty or unreadable; failing closed (denying all requests)")
		} else {
			opts = append(opts, authorization2.NoAuth())
		}
	}

	if a.Authenticate {
		if authenticator == nil {
			return nil, fmt.Errorf("failed to initialize authorizer: no authenticator provided")
		}
		opts = append(opts, authorization2.WithAuthenticator(authenticator))
	}

	return authorization2.New(opts...), nil
}
