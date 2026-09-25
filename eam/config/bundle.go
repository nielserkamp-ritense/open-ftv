package config

import (
	"flag"
	"time"
)

// Bundle contains the configuration variables for sending and receiving policy- and data-bundles.
type Bundle struct {
	BundlePath    string        `env:"BUNDLE_CONFIGS"      flag:"bundle-configs"      desc:"Path for bundle configurations"`
	BundleRecurse bool          `env:"BUNDLE_RECURSE"      flag:"bundle-recurse"      desc:"Search bundle configuration path recursively"`
	BundleTimeout time.Duration `env:"BUNDLE_SEND_TIMEOUT" flag:"bundle-send-timeout" desc:"Timeout for sending a bundle to a PDP (default 1m)"  default:"1m"`
	Workers       int           `env:"BUNDLE_WORKERS"      flag:"bundle-workers"      desc:"Maximum number of worker threads for sending bundles (default #cpu)"`
	StageDelay    time.Duration `env:"BUNDLE_STAGE_DELAY"  flag:"bundle-stage-delay"  desc:"Forced delay between bundle deployment stages"`
}

func (b *Bundle) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&b.BundlePath, "bundle-configs", "", "Path for bundle configurations")
	fs.BoolVar(&b.BundleRecurse, "bundle-recurse", false, "Search bundle configuration path recursively")
	fs.DurationVar(&b.BundleTimeout, "bundle-send-timeout", time.Minute, "Timeout for sending a bundle to a PDP")
	fs.IntVar(&b.Workers, "bundle-workers", 0, "Maximum number of worker threads for sending bundles (default #cpu)")
	fs.DurationVar(&b.StageDelay, "bundle-stage-delay", 0, "Forced delay between bundle deployment stages")
}
