package config

import (
	"flag"
	"time"
)

const defaultBundleRetries = 10

// PDP contains the configuration variables for a generic PDP.
type PDP struct {
	AuthZENMethod   string        `env:"AUTHZEN_METHOD"   flag:"authzen-method"   desc:"HTTP method of the PDP to report in AuthZEN metadata API"`
	AuthZENDomain   string        `env:"AUTHZEN_DOMAIN"   flag:"authzen-domain"   desc:"domain of the PDP to report in AuthZEN metadata API"`
	BundleManager   string        `env:"BUNDLE_MANAGER"   flag:"bundle-manager"   desc:"URI for retrieving the latest bundle from the bundle manager"`
	BundleCert      string        `env:"BUNDLE_CERT"      flag:"bundle-cert"      desc:"TLS Certificate to connect with the bundle manager (https)"`
	BundleKey       string        `env:"BUNDLE_KEY"       flag:"bundle-key"       desc:"TLS Key to connect with the bundle manager (https)"`
	BundleCA        string        `env:"BUNDLE_CA"        flag:"bundle-ca"        desc:"CA certificate key to connect with the bundle manager (mTLS)"`
	BundleAPIKey    string        `env:"BUNDLE_APIKEY"    flag:"bundle-apikey"    desc:"API key to authenticate with the bundle manager ('api-key' header)"`
	BundleHeaders   string        `env:"BUNDLE_HEADERS"   flag:"bundle-headers"   desc:"Optional comma-separated list of headers to send to the bundle manager (a:b,x:y,...)"`
	BundleEncoding  string        `env:"BUNDLE_ENCODING"  flag:"bundle-encoding"  desc:"Optional encoding to receive the latest bundle (gzip or bzip2; default gzip)" default:"gzip"`
	BundleTimeout   time.Duration `env:"BUNDLE_TIMEOUT"   flag:"bundle-timeout"   desc:"Timeout for receiving the latest bundle (default 1m)"                         default:"1m"`
	BundleRetries   int           `env:"BUNDLE_RETRIES"   flag:"bundle-retries"   desc:"Number of times to retry bundle retrieval (default 10)"                       default:"10"`
	RequestMappings string        `env:"REQUEST_MAPPINGS" flag:"request-mappings" desc:"Comma-separated list of attribute/property mappings for PDP requests"`
}

func (p *PDP) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&p.AuthZENMethod, "authzen-method", "", "HTTP method of the PDP to report in AuthZEN metadata API")
	fs.StringVar(&p.AuthZENDomain, "authzen-domain", "", "domain of the PDP to report in AuthZEN metadata API")
	fs.StringVar(&p.BundleManager, "bundle-manager", "", "URI for retrieving the latest bundle from the bundle manager")
	fs.StringVar(&p.BundleCert, "bundle-cert", "", "TLS Certificate to connect with the bundle manager (https)")
	fs.StringVar(&p.BundleKey, "bundle-key", "", "TLS Key to connect with the bundle manager (https)")
	fs.StringVar(&p.BundleCA, "bundle-ca", "", "CA certificate key to connect with the bundle manager (mTLS)")
	fs.StringVar(&p.BundleAPIKey, "bundle-apikey", "", "API key to authenticate with the bundle manager ('api-key' header)")
	fs.StringVar(&p.BundleHeaders, "bundle-headers", "", "Optional comma-separated list of headers to send to the bundle manager (a:b,x:y,...)")
	fs.StringVar(&p.BundleEncoding, "bundle-encoding", "gzip", "Optional encoding to receive the latest bundle (gzip or bzip2)")
	fs.DurationVar(&p.BundleTimeout, "bundle-timeout", time.Minute, "Timeout for receiving the latest bundle")
	fs.IntVar(&p.BundleRetries, "bundle-retries", defaultBundleRetries, "Number of times to retry bundle retrieval")
	fs.StringVar(&p.RequestMappings, "request-mappings", "", "Comma-separated list of attribute/property mappings for PDP requests")
}
