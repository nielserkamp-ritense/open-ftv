package authorization

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authentication"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
)

// NewForController builds an authorizer that decides with the given PDP controller.
//
// With authenticate set, the authorizer first authenticates the caller with authenticator. With failClosedOnEmpty
// set, an empty or unreadable policy store denies all requests instead of allowing them. The extra options let a
// caller add concerns this constructor should not know about, such as the manager passing its principal store.
func NewForController(
	controller pdp.Controller, authenticator authentication.Authenticator, authenticate, failClosedOnEmpty bool, extra ...Option,
) (RequestAuthorizer, error) {
	opts := []Option{
		WithContext(controller.GetContext()),
		WithLogger(controller.GetLogger()),
		WithPEP(controller.GetPEP()),
		WithPDP(controller),
		WithEntityGetter(controller.GetPIP().GetEntity),
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
		if failClosedOnEmpty {
			// Fail closed: do NOT install NoAuth. The PDP evaluates against an empty
			// policy set and denies (default-deny), so a missing/unreadable store locks
			// the API down instead of opening it.
			controller.GetLogger().Warn("authorization: policy store empty or unreadable; failing closed (denying all requests)")
		} else {
			opts = append(opts, NoAuth())
		}
	}

	if authenticate {
		if authenticator == nil {
			return nil, fmt.Errorf("failed to initialize authorizer: no authenticator provided")
		}

		opts = append(opts, WithAuthenticator(authenticator))
	}

	return New(opts...), nil
}
