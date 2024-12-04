package cerbos

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/standards"
)

// Authorize implements the Controller interface.
func (c *controller) Authorize(_ *standards.Request) (*standards.Response, error) {
	resp := &standards.Response{Allowed: true}

	// TODO: determine required policy and execute it

	// all good.
	return resp, nil
}
