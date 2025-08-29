package config

// PDP contains the configuration variables for a generic PDP.
type PDP struct {
	RequestMappings string `yaml:"pdp.request.mappings,omitempty" env:"REQUEST_MAPPINGS" flag:"request-mappings" desc:"Comma-separated list of attribute/property mappings for PDP requests"`
}
