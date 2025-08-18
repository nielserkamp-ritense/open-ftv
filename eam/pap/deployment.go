package pap

import "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"

// NewDeployment creates a new deployment in the store.
func (p *PAP) NewDeployment(description string, manager *bundles.Manager) (*bundles.Deployment, error) {
	p.deployMutex.Lock()
	defer p.deployMutex.Unlock()

	d, err := p.deployer.Generate(description)
	if err != nil {
		return nil, err
	}

	manager.Run(d, p.deployer)
	return d, nil
}

// RestartDeployment checks if a bundle deployment was interrupted and restarts the run if so.
func (p *PAP) RestartDeployment(manager *bundles.Manager) {
	if p.deployer != nil {
		if d, err2 := p.deployer.LastDeployment(); err2 == nil {
			if s := d.Status(); s != bundles.Failed && s != bundles.Completed {
				manager.Run(d, p.deployer)
			}
		}
	}
}

// LastDeployment retrieves the last deployment from the store.
func (p *PAP) LastDeployment() (*bundles.Deployment, error) {
	p.deployMutex.RLock()
	defer p.deployMutex.RUnlock()
	return p.deployer.LastDeployment()
}

// ReadDeployment retrieves a deployment from the store.
func (p *PAP) ReadDeployment(version uint64) (*bundles.Deployment, error) {
	p.deployMutex.RLock()
	defer p.deployMutex.RUnlock()
	return p.deployer.ReadDeployment(version)
}

// ListDeployments retrieves all deployments from the store.
func (p *PAP) ListDeployments() ([]*bundles.Deployment, error) {
	p.deployMutex.RLock()
	defer p.deployMutex.RUnlock()
	return p.deployer.ListDeployments()
}
