package config

import "time"

// Server contains the configuration variables for the main HTTP(S) service.
type Server struct {
	Host         string        `json:"mainHost"             yaml:"svc.host,omitempty"          env:"ADDRESS,HOST"  flag:"address,host,a,h" desc:"Address to use for service (default 0.0.0.0)"      default:"0.0.0.0"`
	Port         uint16        `json:"mainPort"             yaml:"svc.port,omitempty"          env:"PORT"          flag:"port,p"           desc:"Port to use for service (default 8443)"            default:"8443"`
	Domain       string        `json:"mainDomain,omitempty" yaml:"svc.domain,omitempty"        env:"DOMAIN"        flag:"domain,d"         desc:"Domain for the service (should match certificate)"`
	CA           string        `json:"mainCA,omitempty"     yaml:"svc.tls.ca,omitempty"        env:"TLS_CA"        flag:"tls-ca,ca"        desc:"TLS Certificate authority to use for service"`
	Cert         string        `json:"mainCert,omitempty"   yaml:"svc.tls.cert,omitempty"      env:"TLS_CERT"      flag:"tls-cert,cert"    desc:"TLS certificate to use for service"`
	Key          string        `json:"mainKey,omitempty"    yaml:"svc.tls.key,omitempty"       env:"TLS_KEY"       flag:"tls-key,key"      desc:"TLS key authority to use for service"`
	CorsOrigins  string        `json:"mainOrigins"          yaml:"svc.cors.origins,omitempty"  env:"CORS_ORIGINS"  flag:"cors-origins"     desc:"Cors origins (default '*')"                        default:"*"`
	CorsHeaders  string        `json:"mainHeaders"          yaml:"svc.cors.headers,omitempty"  env:"CORS_HEADERS"  flag:"cors-headers"     desc:"Cors headers (default '*')"                        default:"*"`
	ReadTimeout  time.Duration `json:"mainReadTimeout"      yaml:"svc.timeout.read,omitempty"  env:"READ_TIMEOUT"  flag:"read-timeout"     desc:"Read timeout for API requests (default 30s)"       default:"30s"`
	WriteTimeout time.Duration `json:"mainWriteTimeout"     yaml:"svc.timeout.write,omitempty" env:"WRITE_TIMEOUT" flag:"write-timeout"    desc:"Write timeout for API requests (default 30s)"      default:"30s"`
	IdleTimeout  time.Duration `json:"mainIdleTimeout"      yaml:"svc.timeout.idle,omitempty"  env:"IDLE_TIMEOUT"  flag:"idle-timeout"     desc:"Idle timeout for API requests (default 300s)"      default:"300s"`
	MaxBody      int           `json:"mainMaxBody"          yaml:"svc.maxBody,omitempty"       env:"MAX_BODY_SIZE" flag:"max-body"         desc:"Maximum size of request body (default 65536)"      default:"65536"`
}

// InternalServer contains the configuration variables for an internal HTTP(S) service.
type InternalServer struct {
	InternalHost    string        `json:"internalHost,omitempty"   yaml:"svc.internal.host,omitempty"          env:"INTERNAL_HOST"     flag:"internal-host"     desc:"Address to use for internal service (default main host)"`
	InternalPort    uint16        `json:"internalPort"             yaml:"svc.internal.port,omitempty"          env:"INTERNAL_PORT"     flag:"internal-port"     desc:"Port to use for internal service (default 9443)"        default:"9443"`
	InternalDomain  string        `json:"internalDomain,omitempty" yaml:"svc.internal.domain,omitempty"        env:"INTERNAL_DOMAIN"   flag:"internal-domain"   desc:"Domain for internal service (should match certificate)"`
	InternalCA      string        `json:"internalCA,omitempty"     yaml:"svc.internal.tls.ca,omitempty"        env:"INTERNAL_CA"       flag:"internal-ca"       desc:"TLS Certificate authority to use for internal service"`
	InternalCert    string        `json:"internalCert,omitempty"   yaml:"svc.internal.tls.cert,omitempty"      env:"INTERNAL_CERT"     flag:"internal-cert"     desc:"TLS certificate to use for internal service"`
	InternalKey     string        `json:"internalKey,omitempty"    yaml:"svc.internal.tls.key,omitempty"       env:"INTERNAL_KEY"      flag:"internal-key"      desc:"TLS key authority to use for internal service"`
	InternalOrigins string        `json:"internalOrigins"          yaml:"svc.internal.origins,omitempty"       env:"INTERNAL_ORIGINS"  flag:"internal-origins"  desc:"Cors origins for internal service (default '*')"        default:"*"`
	InternalHeaders string        `json:"internalHeaders"          yaml:"svc.internal.headers,omitempty"       env:"INTERNAL_HEADERS"  flag:"internal-headers"  desc:"Cors headers for internal service (default '*')"        default:"*"`
	InternalRead    time.Duration `json:"internalReadTimeout"      yaml:"svc.internal.timeout.read,omitempty"  env:"INTERNAL_READ"     flag:"internal-read"     desc:"Read timeout for internal service (default 30s)"        default:"30s"`
	InternalWrite   time.Duration `json:"internalWriteTimeout"     yaml:"svc.internal.timeout.write,omitempty" env:"INTERNAL_WRITE"    flag:"internal-write"    desc:"Write timeout for internal service (default 30s)"       default:"30s"`
	InternalIdle    time.Duration `json:"internalIdleTimeout"      yaml:"svc.internal.timeout.idle,omitempty"  env:"INTERNAL_IDLE"     flag:"internal-idle"     desc:"Idle timeout for internal service (default 300s)"       default:"300s"`
	InternalMaxBody int           `json:"internalMaxBody"          yaml:"svc.internal.maxBody,omitempty"       env:"INTERNAL_MAX_BODY" flag:"internal-max-body" desc:"Maximum body size for internal service (default 65536)" default:"65536"`
}

// HealthServer contains the configuration variables for a liveness&readiness HTTP service.
type HealthServer struct {
	HealthHost    string        `json:"healthHost,omitempty" yaml:"svc.health.host,omitempty"          env:"HEALTH_HOST"     flag:"health-host"     desc:"Address to use for health service (default main host)"`
	HealthPort    uint16        `json:"healthPort"           yaml:"svc.health.port,omitempty"          env:"HEALTH_PORT"     flag:"health-port"     desc:"Port to use for health service (default 8080)"        default:"8080"`
	HealthOrigins string        `json:"healthOrigins"        yaml:"svc.health.origins,omitempty"       env:"HEALTH_ORIGINS"  flag:"health-origins"  desc:"Cors origins for health service (default '*')"        default:"*"`
	HealthHeaders string        `json:"healthHeaders"        yaml:"svc.health.headers,omitempty"       env:"HEALTH_HEADERS"  flag:"health-headers"  desc:"Cors headers for health service (default '*')"        default:"*"`
	HealthRead    time.Duration `json:"healthReadTimeout"    yaml:"svc.health.timeout.read,omitempty"  env:"HEALTH_READ"     flag:"health-read"     desc:"Read timeout for health service (default 30s)"        default:"30s"`
	HealthWrite   time.Duration `json:"healthWriteTimeout"   yaml:"svc.health.timeout.write,omitempty" env:"HEALTH_WRITE"    flag:"health-write"    desc:"Write timeout for health service (default 30s)"       default:"30s"`
	HealthIdle    time.Duration `json:"healthIdleTimeout"    yaml:"svc.health.timeout.idle,omitempty"  env:"HEALTH_IDLE"     flag:"health-idle"     desc:"Idle timeout for health service (default 300s)"       default:"300s"`
	HealthMaxBody int           `json:"healthMaxBody"        yaml:"svc.health.maxBody,omitempty"       env:"HEALTH_MAX_BODY" flag:"health-max-body" desc:"Maximum body size for health service (default 65536)" default:"256"`
}
