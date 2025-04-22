package mapping

// Option is the function prototype for passing options to mapping functions.
type Option func(m *base)

// WithHeaderKeys sets the header keys to use for a mapping function.
func WithHeaderKeys(keys ...string) Option {
	return func(b *base) {
		b.headerKeys = keys
	}
}
