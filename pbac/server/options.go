package server

import "time"

// ServerOption is the function signature to supply configuration options when instantiating a new service.
type ServerOption func(c *Config)

// WithDefaults sets opinionated default values for the service.
//
// Use this as the first option if you want to override these defaults.
func WithDefaults() ServerOption {
	return func(c *Config) {
		c.LoadDefaults()
	}
}

// WithAppName sets an application name for the service.
func WithAppName(appName string) ServerOption {
	return func(c *Config) {
		c.AppName = appName
	}
}

// WithHostPort sets the host and port for the service to listen on.
func WithHostPort(host string, port uint16) ServerOption {
	return func(c *Config) {
		c.Host = host
		c.Port = port
	}
}

// WithTimeouts sets timeout values for the service.
func WithTimeouts(read, write, idle time.Duration) ServerOption {
	return func(c *Config) {
		c.ReadTimeout = read
		c.WriteTimeout = write
		c.IdleTimeout = idle
	}
}

// WithMaxBody sets the maximum size of request bodies for the service.
func WithMaxBody(maxBody int) ServerOption {
	return func(c *Config) {
		c.MaxBody = maxBody
	}
}

// WithTLS sets TLS parameters for the service.
func WithTLS(certFile, keyFile, caFile string) ServerOption {
	return func(c *Config) {
		c.TLSCert = certFile
		c.TLSKey = keyFile
		c.CA = caFile
	}
}

// WithRecovery turns on automatic panic recovery for the service.
func WithRecovery() ServerOption {
	return func(c *Config) {
		c.RecoverPanics = true
	}
}

// WithSecurity turns on higher security measures for the service.
func WithSecurity() ServerOption {
	return func(c *Config) {
		c.HighSecurity = true
	}
}

// WithCORS sets up CORS support for the service.
func WithCORS(origins, headers string) ServerOption {
	return func(c *Config) {
		c.CorsOrigins = origins
		c.CorsHeaders = headers
	}
}
