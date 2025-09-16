package config

import "time"

// PDP contains the configuration variables for a generic PDP.
type PDP struct {
	AuthZENMethod   string        `yaml:"pdp.authzen.method,omitempty"   env:"AUTHZEN_METHOD"   flag:"authzen-method"   desc:"HTTP method of the PDP to report in AuthZEN metadata API"`
	AuthZENDomain   string        `yaml:"pdp.authzen.domain,omitempty"   env:"AUTHZEN_DOMAIN"   flag:"authzen-domain"   desc:"domain of the PDP to report in AuthZEN metadata API"`
	BundleManager   string        `yaml:"pdp.bundle.manager,omitempty"   env:"BUNDLE_MANAGER"   flag:"bundle-manager"   desc:"URI for retrieving the latest bundle from the bundle manager"`
	BundleHeaders   string        `yaml:"pdp.bundle.headers,omitempty"   env:"BUNDLE_HEADERS"   flag:"bundle-headers"   desc:"Optional comma-separated list of headers to send to the bundle manager"`
	BundleAPIKey    string        `yaml:"pdp.bundle.apiKey,omitempty"    env:"BUNDLE_APIKEY"    flag:"bundle-apikey"    desc:"API key to authenticate with the bundle manager"`
	BundleEncoding  string        `yaml:"pdp.bundle.encoding,omitempty"  env:"BUNDLE_ENCODING"  flag:"bundle-encoding"  desc:"Optional encoding to receive the latest bundle (gzip or bzip2; default gzip)" default:"gzip"`
	BundleTimeout   time.Duration `yaml:"pdp.bundle.timeout,omitempty"   env:"BUNDLE_TIMEOUT"   flag:"bundle-timeout"   desc:"Timeout for receiving the latest bundle (default 1m)"                         default:"1m"`
	BundleRetries   int           `yaml:"pdp.bundle.retries,omitempty"   env:"BUNDLE_RETRIES"   flag:"bundle-retries"   desc:"Number of times to retry bundle retrieval (default 10)"                       default:"10"`
	RequestMappings string        `yaml:"pdp.request.mappings,omitempty" env:"REQUEST_MAPPINGS" flag:"request-mappings" desc:"Comma-separated list of attribute/property mappings for PDP requests"`
}
