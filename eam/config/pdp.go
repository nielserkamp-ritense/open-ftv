package config

import (
	"flag"
	"time"
)

const defaultBundleRetries = 10

// PDP contains the configuration variables for a generic PDP.
type PDP struct {
	AuthZENMethod   string
	AuthZENDomain   string
	BundleManager   string
	BundleCert      string
	BundleKey       string
	BundleCA        string
	BundleAPIKey    string
	BundleHeaders   string
	BundleEncoding  string
	BundleTimeout   time.Duration
	BundleRetries   int
	RequestMappings string
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
