package pap

import (
	policies "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// optionalPrincipal returns nil for an absent attribution, so the field is omitted from the
// response rather than serialized as a principal identifying nobody.
func optionalPrincipal(id string) *policies.Principal {
	if id == "" {
		return nil
	}

	return &policies.Principal{Id: id}
}
