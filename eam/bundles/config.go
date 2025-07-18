package bundles

// Config represents the configuration of a bundle.
//
// The tag is used to determine which elements must go into a bundle.
//
// If a bundle manager is configured with multiple bundles,
// each bundle must have a unique tag.
// Duplicate tags across bundle configurations are not supported.
type Config struct {
	Policies bool      `json:"policies,omitempty" yaml:"policies,omitempty"` // Indicates to bundle the policies with matching tag.
	Data     bool      `json:"data,omitempty"     yaml:"data,omitempty"`     // Indicates to bundle attributes, entities and relations with matching tag.
	Version  bool      `json:"version,omitempty"  yaml:"version,omitempty"`  // Indicates to bundle a new version number.
	Tag      string    `json:"tag"                yaml:"tag"`                // Tag to match against.
	Targets  []*Target `json:"targets"            yaml:"targets"`            // List of target PDPs to send the bundle to.
}
