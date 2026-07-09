package odrl

import "strings"

// Config contains the configuration variables for the ODRL im-/export on a
// PAP or manager. It is embedded in the app configuration structs and follows
// the tag conventions of eam/config.
type Config struct {
	BaseURL     string `yaml:"odrl.baseURL,omitempty" env:"ODRL_BASE_URL" flag:"odrl-base-url" desc:"Public base URL of this PAP, used in generated download-URLs and uids"`
	Annotations string `yaml:"odrl.export.annotations,omitempty" env:"ODRL_EXPORT_ANNOTATIONS" flag:"odrl-export-annotations" desc:"Path to the odrl-export.yaml annotation file (default: <policy store>/odrl-export.yaml)"`
	ExportDir   string `yaml:"odrl.export.dir,omitempty" env:"ODRL_EXPORT_DIR" flag:"odrl-export-dir" desc:"Directory for file-based ODRL export on policy changes (empty disables)"`
	Subscribers string `yaml:"odrl.subscribers,omitempty" env:"ODRL_SUBSCRIBERS" flag:"odrl-subscribers" desc:"Comma separated CloudEvents subscriber URLs to push policy changes to"`
	Source      string `yaml:"odrl.source,omitempty" env:"ODRL_SOURCE" flag:"odrl-source" desc:"CloudEvents source attribute for emitted policy events"`
}

// SubscriberList returns the configured subscriber URLs as a slice.
func (c *Config) SubscriberList() []string {
	if c.Subscribers == "" {
		return nil
	}

	parts := strings.Split(c.Subscribers, ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
