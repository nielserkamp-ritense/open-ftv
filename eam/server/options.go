package server

import "time"

// Option is the function signature to supply configuration options when instantiating a new service.
type Option func(c *Config)

// WithDefaults sets opinionated default values for the service.
//
// Use this as the first option if you want to override these defaults.
func WithDefaults() Option {
	return func(c *Config) {
		c.LoadDefaults()
	}
}

// WithAppName sets an application name for the service.
func WithAppName(appName string) Option {
	return func(c *Config) {
		c.AppName = appName
	}
}

// WithHostPort sets the host and port for the service to listen on.
func WithHostPort(host string, port uint16) Option {
	return func(c *Config) {
		c.Host = host
		c.Port = port
	}
}

// WithTimeouts sets timeout values for the service.
func WithTimeouts(read, write, idle time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = read
		c.WriteTimeout = write
		c.IdleTimeout = idle
	}
}

// WithMaxBody sets the maximum size of request bodies for the service.
func WithMaxBody(maxBody int) Option {
	return func(c *Config) {
		c.MaxBody = maxBody
	}
}

// WithTLS sets TLS parameters for the service.
func WithTLS(caFile, certFile, keyFile string) Option {
	return func(c *Config) {
		c.CA = caFile
		c.TLSCert = certFile
		c.TLSKey = keyFile
	}
}

// WithMutualTLS switches to mutual TLS for the service.
//
// The service will require and verify a certificate from clients.
func WithMutualTLS() Option {
	return func(c *Config) {
		c.MutualTLS = true
	}
}

// WithRecovery turns on automatic panic recovery for the service.
func WithRecovery() Option {
	return func(c *Config) {
		c.RecoverPanics = true
	}
}

// WithSecurity turns on higher security measures for the service.
func WithSecurity() Option {
	return func(c *Config) {
		c.HighSecurity = true
	}
}

// WithCORS sets up CORS support for the service.
func WithCORS(origins, headers string) Option {
	return func(c *Config) {
		c.CorsOrigins = origins
		c.CorsHeaders = headers
	}
}
