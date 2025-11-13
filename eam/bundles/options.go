package bundles

import "time"

// Option is the function signature for passing options when instantiating a new bundle deployment manager.
type Option func(*Manager)

// WithConfig adds the path where the bundle configuration can be found.
func WithConfig(path string, recurse bool) Option {
	return func(m *Manager) {
		m.path = path
		m.recurse = recurse
	}
}

// WithPolicyHandler adds the interface to work with policies.
func WithPolicyHandler(lister PolicyHandler) Option {
	return func(m *Manager) {
		m.policies = lister
	}
}

// WithDataHandler adds the interface to work with attributes, entities and relations.
func WithDataHandler(iterator DataHandler) Option {
	return func(m *Manager) {
		m.data = iterator
	}
}

// MaxWorkers sets the maximum number of workers for sending bundles.
func MaxWorkers(workers int) Option {
	return func(m *Manager) {
		m.workers = workers
	}
}

// WithStageDelay sets the forced delay between bundle deployment stages.
func WithStageDelay(delay time.Duration) Option {
	return func(m *Manager) {
		m.stageDelay = delay
	}
}

// BundleTimeout sets the timeout for sending a bundle to a PDP.
func BundleTimeout(timeout time.Duration) Option {
	return func(m *Manager) {
		m.bundleTimeout = timeout
	}
}

// BootstrapDeployment sets the flag to create an initial bootstrap deployment.
func BootstrapDeployment() Option {
	return func(m *Manager) {
		m.bootstrap = true
	}
}
