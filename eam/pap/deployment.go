package pap

import (
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
)

// NewDeployment creates a new deployment in the store.
func (p *PAP) NewDeployment(title, description string, manager *bundles.Manager, user identity.Principal) (*bundles.Deployment, error) {
	p.deployMutex.Lock()
	defer p.deployMutex.Unlock()

	// Generate still takes a plain string; convert at this boundary.
	d, err := p.bundleDB.Generate(p.ctx, title, description, user.DisplayName())
	if err != nil {
		return nil, err
	}

	manager.Run(d, p.bundleDB, user)
	return d, nil
}

// RestartDeployment checks if a bundle deployment was interrupted and restarts the run if so.
func (p *PAP) RestartDeployment(manager *bundles.Manager) {
	if p.bundleDB != nil {
		if d, err2 := p.bundleDB.LastDeployment(p.ctx); err2 == nil && d != nil {
			if s := d.Status(); s != bundles.Failed && s != bundles.Completed {
				manager.Run(d, p.bundleDB, identity.NewSystemPrincipal())
			}
		}
	}
}

// LastDeployment retrieves the last deployment from the store.
func (p *PAP) LastDeployment() (*bundles.Deployment, error) {
	p.deployMutex.RLock()
	defer p.deployMutex.RUnlock()
	return p.bundleDB.LastDeployment(p.ctx)
}

// ReadDeployment retrieves a deployment from the store.
func (p *PAP) ReadDeployment(version uint64) (*bundles.Deployment, error) {
	p.deployMutex.RLock()
	defer p.deployMutex.RUnlock()
	return p.bundleDB.ReadDeployment(p.ctx, version)
}

// ListDeployments retrieves all deployments from the store.
func (p *PAP) ListDeployments() ([]*bundles.Deployment, error) {
	p.deployMutex.RLock()
	defer p.deployMutex.RUnlock()
	return p.bundleDB.ListDeployments(p.ctx)
}
