package openfga

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/shared"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(_ *shared.Request) (*shared.Response, error) {
	resp := &shared.Response{Allowed: true}

	// TODO: determine required policy and execute it

	// all good.
	return resp, nil
}
