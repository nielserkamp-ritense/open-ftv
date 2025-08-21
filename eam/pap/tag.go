package pap

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// ListTags returns a sorted list of all tags in the database.
func (p *PAP) ListTags() (out policies.Tags, err error) {
	return p.tagDB.ListTags(p.ctx)
}
