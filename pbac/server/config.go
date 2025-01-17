// Package config contains the service configuration.
package server

import "time"

// Config contains the configurable parameters for a service.
type Config struct {
	Port          uint16        // server port to listen on.
	MaxBody       int           // maximum size of request bodies.
	AppName       string        // application name.
	Host          string        // host address to listen on.
	CA            string        // path to Certificate Authority file.
	TLSCert       string        // path to TLS certificate file.
	TLSKey        string        // path to TLS key file.
	ReadTimeout   time.Duration // timeout for reads.
	WriteTimeout  time.Duration // timeout for writes.
	IdleTimeout   time.Duration // idle timeout for open connections.
	RecoverPanics bool          // switch on automatic panic recovery.
	HighSecurity  bool          // enable security measures.
	CorsOrigins   string        // allowed CORS origins.
	CorsHeaders   string        // allowed CORS headers.
}

// LoadDefaults loads opinionated default values into the service configuration.
//
// host&port          : 0.0.0.0:8080.
// read/write timeouts: 10 seconds.
// idle timeout       : 5 minutes.
// max request body   : 65536 bytes.
func (c *Config) LoadDefaults() {
	c.Host = "0.0.0.0"
	c.Port = 8080
	c.MaxBody = 65536
	c.ReadTimeout = time.Second * 10
	c.WriteTimeout = time.Second * 10
	c.IdleTimeout = time.Minute * 5
}

// LoadOptions loads the given options into the service configuration.
func (c *Config) LoadOptions(opts ...ServerOption) {
	for i := range opts {
		opts[i](c)
	}
}
