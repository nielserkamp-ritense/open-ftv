package pap

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// CreateTag adds a tag to cache/storage.
//
// An error is returned if the policy-id already exists.
func (p *PAP) CreateTag(in *oas.Tag, user identity.Principal) (*oas.Tag, error) {
	return p.tagDB.CreateTag(p.ctx, user, in)
}

// ReadTag retrieves a tag from cache/storage.
//
// An error is returned if the policy-id doesn't exist.
func (p *PAP) ReadTag(id string) (out *oas.Tag, lastIndex uint64, err error) {
	return p.tagDB.ReadTag(p.ctx, id)
}

// UpdateTag modifies a tag in cache/storage with a newer version.
//
// An error is returned if the policy-id doesn't exist.
func (p *PAP) UpdateTag(prev *oas.Tag, lastIndex uint64, in *oas.Tag, user identity.Principal) (out *oas.Tag, err error) {
	return p.tagDB.UpdateTag(p.ctx, user, prev, lastIndex, in)
}

// DeleteTag removes a tag from cache/storage.
//
// An error is returned if the policy key doesn't exist.
func (p *PAP) DeleteTag(prev *oas.Tag, lastIndex uint64, user identity.Principal) (out *oas.Tag, err error) {
	return p.tagDB.DeleteTag(p.ctx, user, prev, lastIndex)
}

// ListTags returns a sorted list of all tags in the database.
func (p *PAP) ListTags() (out oas.Tags, err error) {
	return p.tagDB.ListTags(p.ctx)
}
