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
	Host         string
	Port         uint16
	Domain       string
	CA           string
	Cert         string
	Key          string
	CorsOrigins  string
	CorsHeaders  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// InternalServer contains the configuration variables for an internal HTTP(S) service.
type InternalServer struct {
	InternalHost    string
	InternalPort    uint16
	InternalDomain  string
	InternalCA      string
	InternalCert    string
	InternalKey     string
	InternalOrigins string
	InternalHeaders string
	InternalRead    time.Duration
	InternalWrite   time.Duration
	InternalIdle    time.Duration
}

// HealthServer contains the configuration variables for a liveness&readiness HTTP service.
type HealthServer struct {
	HealthHost    string
	HealthPort    uint16
	HealthOrigins string
	HealthHeaders string
	HealthRead    time.Duration
	HealthWrite   time.Duration
	HealthIdle    time.Duration
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
