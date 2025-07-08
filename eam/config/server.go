package config

import "time"

// Server contains the configuration variables for an HTTP service.
type Server struct {
	Host         string        `yaml:"svc.host,omitempty" env:"ADDRESS,HOST" flag:"address,a" default:"0.0.0.0" desc:"Address to use for service"`
	Port         uint16        `yaml:"svc.port,omitempty" env:"PORT" flag:"port,p" default:"8443" desc:"Port to use for https service"`
	CA           string        `yaml:"svc.tls.ca,omitempty" env:"TLS_CA" flag:"tls-ca,ca" desc:"TLS Certificate authority to use for https service"`
	Cert         string        `yaml:"svc.tls.cert,omitempty" env:"TLS_CERT" flag:"tls-cert,cert" desc:"TLS certificate to use for https service"`
	Key          string        `yaml:"svc.tls.key,omitempty" env:"TLS_KEY" flag:"tls-key,key" desc:"TLS key authority to use for https service"`
	ReadTimeout  time.Duration `yaml:"svc.timeout.read,omitempty" env:"READ_TIMEOUT" flag:"read-timeout" default:"30s" desc:"Read timeout for API requests"`
	WriteTimeout time.Duration `yaml:"svc.timeout.write,omitempty" env:"WRITE_TIMEOUT" flag:"write-timeout" default:"30s" desc:"Write timeout for API requests"`
	IdleTimeout  time.Duration `yaml:"svc.timeout.idle,omitempty" env:"IDLE_TIMEOUT" flag:"idle-timeout" default:"300s" desc:"Idle timeout for API requests"`
	MaxBody      int           `yaml:"svc.maxBody,omitempty" env:"MAX_BODY_SIZE" flag:"max-body" default:"65536" desc:"Maximum size of request body"`
	CorsOrigins  string        `yaml:"svc.cors.origins,omitempty" env:"CORS_ORIGINS" flag:"cors-origins" default:"*" desc:"Cors origins"`
	CorsHeaders  string        `yaml:"svc.cors.headers,omitempty" env:"CORS_HEADERS" flag:"cors-headers" default:"*" desc:"Cors headers"`
}
