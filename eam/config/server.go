package config

import (
	"flag"
	"time"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/envflag"
)

const (
	defaultPort         = 8443
	defaultInternalPort = 9443
	defaultHealthPort   = 8080
	defaultTimeout      = 30 * time.Second
	defaultIdleTimeout  = 300 * time.Second
)

// Server contains the configuration variables for the main HTTP(S) service.
type Server struct {
	Host         string        `env:"ADDRESS,HOST"  flag:"address,host,a,h" desc:"Address to use for service (default 0.0.0.0)"      default:"0.0.0.0"`
	Port         uint16        `env:"PORT"          flag:"port,p"           desc:"Port to use for service (default 8443)"            default:"8443"`
	Domain       string        `env:"DOMAIN"        flag:"domain,d"         desc:"Domain for the service (should match certificate)"`
	CA           string        `env:"TLS_CA"        flag:"tls-ca,ca"        desc:"TLS Certificate authority to use for service"`
	Cert         string        `env:"TLS_CERT"      flag:"tls-cert,cert"    desc:"TLS certificate to use for service"`
	Key          string        `env:"TLS_KEY"       flag:"tls-key,key"      desc:"TLS key authority to use for service"`
	CorsOrigins  string        `env:"CORS_ORIGINS"  flag:"cors-origins"     desc:"Cors origins (default '*')"                        default:"*"`
	CorsHeaders  string        `env:"CORS_HEADERS"  flag:"cors-headers"     desc:"Cors headers (default '*')"                        default:"*"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT"  flag:"read-timeout"     desc:"Read timeout for API requests (default 30s)"       default:"30s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" flag:"write-timeout"    desc:"Write timeout for API requests (default 30s)"      default:"30s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT"  flag:"idle-timeout"     desc:"Idle timeout for API requests (default 300s)"      default:"300s"`
}

// InternalServer contains the configuration variables for an internal HTTP(S) service.
type InternalServer struct {
	InternalHost    string        `env:"INTERNAL_HOST"     flag:"internal-host"     desc:"Address to use for internal service (default main host)"`
	InternalPort    uint16        `env:"INTERNAL_PORT"     flag:"internal-port"     desc:"Port to use for internal service (default 9443)"        default:"9443"`
	InternalDomain  string        `env:"INTERNAL_DOMAIN"   flag:"internal-domain"   desc:"Domain for internal service (should match certificate)"`
	InternalCA      string        `env:"INTERNAL_CA"       flag:"internal-ca"       desc:"TLS Certificate authority to use for internal service"`
	InternalCert    string        `env:"INTERNAL_CERT"     flag:"internal-cert"     desc:"TLS certificate to use for internal service"`
	InternalKey     string        `env:"INTERNAL_KEY"      flag:"internal-key"      desc:"TLS key authority to use for internal service"`
	InternalOrigins string        `env:"INTERNAL_ORIGINS"  flag:"internal-origins"  desc:"Cors origins for internal service (default '*')"        default:"*"`
	InternalHeaders string        `env:"INTERNAL_HEADERS"  flag:"internal-headers"  desc:"Cors headers for internal service (default '*')"        default:"*"`
	InternalRead    time.Duration `env:"INTERNAL_READ"     flag:"internal-read"     desc:"Read timeout for internal service (default 30s)"        default:"30s"`
	InternalWrite   time.Duration `env:"INTERNAL_WRITE"    flag:"internal-write"    desc:"Write timeout for internal service (default 30s)"       default:"30s"`
	InternalIdle    time.Duration `env:"INTERNAL_IDLE"     flag:"internal-idle"     desc:"Idle timeout for internal service (default 300s)"       default:"300s"`
}

// HealthServer contains the configuration variables for a liveness&readiness HTTP service.
type HealthServer struct {
	HealthHost    string        `env:"HEALTH_HOST"     flag:"health-host"     desc:"Address to use for health service (default main host)"`
	HealthPort    uint16        `env:"HEALTH_PORT"     flag:"health-port"     desc:"Port to use for health service (default 8080)"        default:"8080"`
	HealthOrigins string        `env:"HEALTH_ORIGINS"  flag:"health-origins"  desc:"Cors origins for health service (default '*')"        default:"*"`
	HealthHeaders string        `env:"HEALTH_HEADERS"  flag:"health-headers"  desc:"Cors headers for health service (default '*')"        default:"*"`
	HealthRead    time.Duration `env:"HEALTH_READ"     flag:"health-read"     desc:"Read timeout for health service (default 30s)"        default:"30s"`
	HealthWrite   time.Duration `env:"HEALTH_WRITE"    flag:"health-write"    desc:"Write timeout for health service (default 30s)"       default:"30s"`
	HealthIdle    time.Duration `env:"HEALTH_IDLE"     flag:"health-idle"     desc:"Idle timeout for health service (default 300s)"       default:"300s"`
}

func (s *Server) RegisterFlags(fs *flag.FlagSet) { //nolint:dupl // the internal service takes the same settings under its own names
	fs.StringVar(&s.Host, "host", "0.0.0.0", "Address to use for service")
	envflag.Uint16Var(fs, &s.Port, "port", defaultPort, "Port to use for service")
	fs.StringVar(&s.Domain, "domain", "", "Domain for the service (should match certificate)")
	fs.StringVar(&s.CA, "tls-ca", "", "TLS Certificate authority to use for service")
	fs.StringVar(&s.Cert, "tls-cert", "", "TLS certificate to use for service")
	fs.StringVar(&s.Key, "tls-key", "", "TLS key authority to use for service")
	fs.StringVar(&s.CorsOrigins, "cors-origins", "*", "Cors origins")
	fs.StringVar(&s.CorsHeaders, "cors-headers", "*", "Cors headers")
	fs.DurationVar(&s.ReadTimeout, "read-timeout", defaultTimeout, "Read timeout for API requests")
	fs.DurationVar(&s.WriteTimeout, "write-timeout", defaultTimeout, "Write timeout for API requests")
	fs.DurationVar(&s.IdleTimeout, "idle-timeout", defaultIdleTimeout, "Idle timeout for API requests")
}

func (s *InternalServer) RegisterFlags(fs *flag.FlagSet) { //nolint:dupl // the main service takes the same settings under its own names
	fs.StringVar(&s.InternalHost, "internal-host", "", "Address to use for internal service (default main host)")
	envflag.Uint16Var(fs, &s.InternalPort, "internal-port", defaultInternalPort, "Port to use for internal service")
	fs.StringVar(&s.InternalDomain, "internal-domain", "", "Domain for internal service (should match certificate)")
	fs.StringVar(&s.InternalCA, "internal-ca", "", "TLS Certificate authority to use for internal service")
	fs.StringVar(&s.InternalCert, "internal-cert", "", "TLS certificate to use for internal service")
	fs.StringVar(&s.InternalKey, "internal-key", "", "TLS key authority to use for internal service")
	fs.StringVar(&s.InternalOrigins, "internal-origins", "*", "Cors origins for internal service")
	fs.StringVar(&s.InternalHeaders, "internal-headers", "*", "Cors headers for internal service")
	fs.DurationVar(&s.InternalRead, "internal-read", defaultTimeout, "Read timeout for internal service")
	fs.DurationVar(&s.InternalWrite, "internal-write", defaultTimeout, "Write timeout for internal service")
	fs.DurationVar(&s.InternalIdle, "internal-idle", defaultIdleTimeout, "Idle timeout for internal service")
}

func (s *HealthServer) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&s.HealthHost, "health-host", "", "Address to use for health service (default main host)")
	envflag.Uint16Var(fs, &s.HealthPort, "health-port", defaultHealthPort, "Port to use for health service")
	fs.StringVar(&s.HealthOrigins, "health-origins", "*", "Cors origins for health service")
	fs.StringVar(&s.HealthHeaders, "health-headers", "*", "Cors headers for health service")
	fs.DurationVar(&s.HealthRead, "health-read", defaultTimeout, "Read timeout for health service")
	fs.DurationVar(&s.HealthWrite, "health-write", defaultTimeout, "Write timeout for health service")
	fs.DurationVar(&s.HealthIdle, "health-idle", defaultIdleTimeout, "Idle timeout for health service")
}
