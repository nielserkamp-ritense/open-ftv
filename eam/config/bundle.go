package config

// Bundle contains the configuration variables for sending and receiving policy- and data-bundles.
type Bundle struct {
	Path string `yaml:"bundles.configPath,omitempty" env:"BUNDLE_CONFIGS" flag:"bundle-configs" desc:"Path for bundle configurations"`
}
