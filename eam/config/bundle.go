package config

import "time"

// Bundle contains the configuration variables for sending and receiving policy- and data-bundles.
type Bundle struct {
	BundlePath    string        `yaml:"bundle.path,omitempty"        env:"BUNDLE_CONFIGS"      flag:"bundle-configs"                   desc:"Path for bundle configurations"`
	BundleRecurse bool          `yaml:"bundle.recurse,omitempty"     env:"BUNDLE_RECURSE"      flag:"bundle-recurse"                   desc:"Search bundle configuration path recursively"`
	BundleTimeout time.Duration `yaml:"bundle.sendTimeout,omitempty" env:"BUNDLE_SEND_TIMEOUT" flag:"bundle-send-timeout" default:"1m" desc:"Timeout for sending a bundle to a PDP (default 1m)"`
	Workers       int           `yaml:"bundle.workers,omitempty"     env:"BUNDLE_WORKERS"      flag:"bundle-workers"                   desc:"Maximum number of worker threads for sending bundles (default #cpu)"`
	StageDelay    time.Duration `yaml:"bundle.stageDelay,omitempty"  env:"BUNDLE_STAGE_DELAY"  flag:"bundle-stage-delay"               desc:"Forced delay between bundle deployment stages"`
}
