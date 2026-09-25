package config

import (
	"flag"
	"time"
)

// Bundle contains the configuration variables for sending and receiving policy- and data-bundles.
type Bundle struct {
	BundlePath    string
	BundleRecurse bool
	BundleTimeout time.Duration
	Workers       int
	StageDelay    time.Duration
}

func (b *Bundle) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&b.BundlePath, "bundle-configs", "", "Path for bundle configurations")
	fs.BoolVar(&b.BundleRecurse, "bundle-recurse", false, "Search bundle configuration path recursively")
	fs.DurationVar(&b.BundleTimeout, "bundle-send-timeout", time.Minute, "Timeout for sending a bundle to a PDP")
	fs.IntVar(&b.Workers, "bundle-workers", 0, "Maximum number of worker threads for sending bundles (default #cpu)")
	fs.DurationVar(&b.StageDelay, "bundle-stage-delay", 0, "Forced delay between bundle deployment stages")
}
