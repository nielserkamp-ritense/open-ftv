package pap

import (
	"fmt"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
)

// Create adds a policy to cache/storage.
//
// An error is returned if the policy-id already exists.
func (p *PAP) Create(in *models.Policy, user identity.Principal) (out *models.Policy, err error) {
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

// ReadAudit retrieves the audit-log for a policy from cache/storage.
func (p *PAP) ReadAudit(id string) ([]oas.AuditEntry, error) {
	return p.policyDB.ReadPolicyAudit(p.ctx, id)
}

// ReadDeployments retrieves the deployment-log for a policy from cache/storage.
func (p *PAP) ReadDeployments(id string) ([]oas.UsageData, error) {
	return p.policyDB.ReadPolicyDeployments(p.ctx, id)
}

// ReadVersions retrieves the versions for a policy from cache/storage.
func (p *PAP) ReadVersions(id string) (oas.PolicyVersions, error) {
	return p.policyDB.ReadPolicyVersions(p.ctx, id)
}

// ReadVersion retrieves a specific version for a policy from cache/storage.
func (p *PAP) ReadVersion(id string, version int) (*oas.PolicyVersion, error) {
	return p.policyDB.ReadPolicyVersion(p.ctx, id, version)
}

// RestoreVersion restores a specific version of a policy as the current concept.
func (p *PAP) RestoreVersion(id string, version int, user identity.Principal) (*models.Policy, error) {
	polOld, err := p.policyDB.ReadPolicyVersion(p.ctx, id, version)
	if err != nil {
		return nil, err
	}

	pol, lastIndex, err2 := p.policyDB.ReadPolicy(p.ctx, id)
	if err2 != nil {
		return nil, err2
	}

	pol = pol.RestoreFrom(polOld)
	return p.Update(pol, lastIndex, pol, user)
}

// Update modifies a policy in cache/storage with a newer version.
//
// An error is returned if the policy-id doesn't exist.
func (p *PAP) Update(prev *models.Policy, lastIndex uint64, in *models.Policy, user identity.Principal) (out *models.Policy, err error) {
	if out, err = p.policyDB.UpdatePolicy(p.ctx, user, prev, lastIndex, in); err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.PolicyReplaced, out.Key())
	}
	return
}

// UpdateStatus updates the status of a policy in cache/storage.
//
// An error is returned if the policy key doesn't exist or the status update is not allowed.
func (p *PAP) UpdateStatus(prev *models.Policy, lastIndex uint64, status models.Status, user identity.Principal) (out *models.Policy, err error) {
	switch prev.Status() {
	case models.StatusConcept:
		if status != models.StatusAccepted {
			return nil, fmt.Errorf("invalid status change from %s to %s", prev.Status().String(), status.String())
		}
	case models.StatusAccepted:
		if status != models.StatusConcept {
			return nil, fmt.Errorf("invalid status change from %s to %s", prev.Status().String(), status.String())
		}
	case models.StatusDeployed:
		return nil, fmt.Errorf("current status cannot be changed: %s", prev.Status().String())
	}

	return p.policyDB.UpdatePolicy(p.ctx, user, prev, lastIndex, prev.WithStatus(status))
}

// Delete removes a policy from cache/storage.
//
// An error is returned if the policy key doesn't exist.
func (p *PAP) Delete(prev *models.Policy, lastIndex uint64, user identity.Principal) (out *models.Policy, err error) {
	if out, err = p.policyDB.DeletePolicy(p.ctx, user, prev, lastIndex); err == nil && out != nil && p.eventSinks != nil {
		p.sendEvent(models.PolicyRemoved, out.Key())
	}
	return
}

// ReplaceAll removes all policies from cache/storage and adds the given list.
func (p *PAP) ReplaceAll(list []*models.Policy, user identity.Principal) error {
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
