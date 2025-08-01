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

// WithPolicyLister adds the function to list all policies for possible inclusion in bundles.
func WithPolicyLister(lister PolicyLister) Option {
	return func(m *Manager) {
		m.policies = lister
	}
}

// WithAttributeLister adds the function to list all attributes for possible inclusion in bundles.
func WithAttributeLister(iterator AttributeLister) Option {
	return func(m *Manager) {
		m.attributes = iterator
	}
}

// WithEntityLister adds the function to list all entities for possible inclusion in bundles.
func WithEntityLister(iterator EntityLister) Option {
	return func(m *Manager) {
		m.entities = iterator
	}
}

// WithRelationLister adds the function to list all relations for possible inclusion in bundles.
func WithRelationLister(iterator RelationLister) Option {
	return func(m *Manager) {
		m.relations = iterator
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
