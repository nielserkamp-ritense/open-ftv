package adl

// Option represents the function signature for supplying options when instantiating a new Authorization Decision ADL.
type Option func(*ADL)

// WithBundleVersion adds the initial bundle version to the ADL.
func WithBundleVersion(version uint64) Option {
	return func(l *ADL) {
		l.bundleVersion = version
	}
}

// WithInformation adds initial information to the ADL.
func WithInformation(information map[string]any) Option {
	return func(l *ADL) {
		l.information = information
	}
}

// WithEngine adds the initial policy-engine details to the ADL.
func WithEngine(engine map[string]any) Option {
	return func(l *ADL) {
		l.engine = engine
	}
}

// WithResource sets the Logius ADL resource, identifying the producer of the
// log record (the system, application, or environment in which the PDP
// evaluated the authorization decision).
func WithResource(resource map[string]any) Option {
	return func(l *ADL) {
		l.resource = resource
	}
}
