package config

import "strings"

// ADL contains the configuration variables for the Authorization Decision Log (ADL).
//
// With an application environment prefix such as FTV_PDP_ or FSC_AUTH_, the resulting
// environment variables are e.g. FTV_PDP_ADL_ENABLED, FTV_PDP_ADL_LEVEL, and so on.
type ADL struct {
	Enabled    bool   `yaml:"adl.enabled,omitempty" env:"ADL_ENABLED" flag:"adl-enabled" desc:"Enable ADL-conformant decision logging"`
	Level      int    `yaml:"adl.level,omitempty" env:"ADL_LEVEL" flag:"adl-level" default:"1" desc:"ADL level of detail (1-4)"`
	Path       string `yaml:"adl.path,omitempty" env:"ADL_PATH" flag:"adl-path" desc:"File path for the durable ADL write-ahead log (JSONL)"`
	OTLPURL    string `yaml:"adl.otlp.url,omitempty" env:"ADL_OTLP_URL" flag:"adl-otlp-url" desc:"OTLP endpoint for asynchronous ADL flushing (or stdout, stderr, slog)"`
	Resource   string `yaml:"adl.resource,omitempty" env:"ADL_RESOURCE" flag:"adl-resource" desc:"Producer identification as comma-separated key=value pairs (e.g. service.name=pdp,env=prod)"`
	ConfigHash string `yaml:"adl.configHash,omitempty" env:"ADL_CONFIG_HASH" flag:"adl-config-hash" desc:"Hash or version of the deployed configuration, recorded at ADL level 4"`
	Insecure   bool   `yaml:"adl.insecure,omitempty" env:"ADL_INSECURE" flag:"adl-insecure" desc:"Allow cleartext (http://) ADL sink endpoints; the standard requires TLS, so this is an explicit opt-out for development or closed networks only (loopback is always allowed)"`
}

// ResourceMap parses the Resource variable ("k1=v1,k2=v2") into the resource object that
// identifies the producer of ADL records. The given defaults are used for any key not
// set explicitly; typically the application supplies at least a service name.
func (a *ADL) ResourceMap(defaults map[string]any) map[string]any {
	out := make(map[string]any, len(defaults)+4)
	for k, v := range defaults {
		out[k] = v
	}

	for _, pair := range strings.Split(a.Resource, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		if k, v, ok := strings.Cut(pair, "="); ok && strings.TrimSpace(k) != "" {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}

	return out
}
