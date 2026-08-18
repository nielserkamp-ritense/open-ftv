package pip

import (
	attributes "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
)

// optionalPrincipal returns nil for an absent attribution, so the field is omitted from the
// response rather than serialized as a principal identifying nobody.
func optionalPrincipal(id string) *attributes.Principal {
	if id == "" {
		return nil
	}

	return &attributes.Principal{Id: id}
}
