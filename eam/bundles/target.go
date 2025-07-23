package bundles

// Target represents the details of a target PDP to receive bundles.
type Target struct {
	URI      string   `json:"uri"                   yaml:"uri"`                   // URI of the PDP.
	CA       string   `json:"ca,omitempty"          yaml:"ca,omitempty"`          // CA certificate file for TLS.
	Cert     string   `json:"cert,omitempty"        yaml:"cert,omitempty"`        // Certificate file for TLS.
	Key      string   `json:"key,omitempty"         yaml:"key,omitempty"`         // Private key file for TLS.
	APIKey   string   `json:"apiKey,omitempty"      yaml:"apiKey,omitempty"`      // API key to authenticate with the PDP.
	Headers  []string `json:"headers,omitempty"     yaml:"headers,omitempty"`     // Additional headers to send.
	Compress string   `json:"compression,omitempty" yaml:"compression,omitempty"` // Compression type for the bundle.
}
