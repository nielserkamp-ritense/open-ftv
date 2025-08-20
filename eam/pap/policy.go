package pap

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"

// Create adds a policy to cache/storage.
//
// An error is returned if the policy-id already exists.
func (p *PAP) Create(in *models.Policy, user string) (out *models.Policy, err error) {
	if out, err = p.policyDB.CreatePolicy(p.ctx, user, in); err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.PolicyAdded, out.Key())
	}
	return
}

// Read retrieves a policy from cache/storage.
//
// An error is returned if the policy-id doesn't exist.
func (p *PAP) Read(id string) (out *models.Policy, lastIndex uint64, err error) {
	return p.policyDB.ReadPolicy(p.ctx, id)
}

// Update modifies a policy in cache/storage with a newer version.
//
// An error is returned if the policy-id doesn't exist.
func (p *PAP) Update(prev *models.Policy, lastIndex uint64, in *models.Policy, user string) (out *models.Policy, err error) {
	if out, err = p.policyDB.UpdatePolicy(p.ctx, user, prev, lastIndex, in); err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.PolicyReplaced, out.Key())
	}
	return
}

// Delete removes a policy from cache/storage.
//
// An error is returned if the policy key doesn't exist.
func (p *PAP) Delete(prev *models.Policy, lastIndex uint64, user string) (out *models.Policy, err error) {
	if out, err = p.policyDB.DeletePolicy(p.ctx, user, prev, lastIndex); err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.PolicyRemoved, out.Key())
	}
	return
}

// ReplaceAll removes all policies from cache/storage and adds the given list.
func (p *PAP) ReplaceAll(list []*models.Policy, user string) error {
	// delete all existing policies.
	old, _ := p.List("")
	for _, policy := range old {
		_, ix, err := p.Read(policy.ID())
		if err != nil {
			return err
		}
		if _, err = p.Delete(policy, ix, user); err != nil {
			return err
		}
	}

	// add all given policies.
	for i := range list {
		if _, err := p.Create(list[i], user); err != nil {
			return err
		}
	}

	return nil
}

// List returns a sorted list of all cached/stored policies.
//
// If the optional language parameter is supplied,
// the function lists all policies with that language.
// Otherwise, all policies, regardless of language, will be listed.
func (p *PAP) List(language string) (out []*models.Policy, err error) {
	return p.policyDB.ListPolicies(p.ctx, language)
}

// Iterate calls the given closure for all policies in the store.
//
// Note that the PAP is locked during the iteration,
// so make sure the given closure does not block or take a long time to process.
func (p *PAP) Iterate(f models.PolicyIterator) {
	list, _ := p.List("")
	for i := range list {
		f(list[i])
	}
}
