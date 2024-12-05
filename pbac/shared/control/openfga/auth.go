package openfga

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared/control"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(_ *control.Request) (*control.Response, error) {
	resp := &control.Response{Allowed: true}

	// TODO: determine required policy and execute it

	// all good.
	return resp, nil
}
