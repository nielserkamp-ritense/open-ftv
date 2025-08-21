package pap

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"

// NewDeployment creates a new deployment in the store.
func (p *PAP) NewDeployment(description string, manager *bundles.Manager, user string) (*bundles.Deployment, error) {
	p.deployMutex.Lock()
	defer p.deployMutex.Unlock()

	d, err := p.bundleDB.Generate(p.ctx, description, user)
	if err != nil {
		return nil, err
	}

	manager.Run(d, p.bundleDB)
	return d, nil
}

// RestartDeployment checks if a bundle deployment was interrupted and restarts the run if so.
func (p *PAP) RestartDeployment(manager *bundles.Manager) {
	if p.bundleDB != nil {
		if d, err2 := p.bundleDB.LastDeployment(p.ctx); err2 == nil && d != nil {
			if s := d.Status(); s != bundles.Failed && s != bundles.Completed {
				manager.Run(d, p.bundleDB)
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
