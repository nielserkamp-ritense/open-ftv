package cerbos

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(_ *components.Request) (*components.Response, error) {
	resp := &components.Response{Allowed: true}

	// TODO: determine required policy and execute it

	// all good.
	return resp, nil
}
