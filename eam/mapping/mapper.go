package mapping

import (
	"fmt"
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

// Mapper is the function signature for mapping specific attributes from known locations into a preferred location.
//
// E,g, it can be used to map a specific attribute value as the principal, action or resource.
// This is useful to simplify policy selection based on attributes instead of the supplied principal, action and resource triplet.
type Mapper func(parc *models.PARC, opts ...Option) *models.PARC

// registry holds the mappers registered by plugin modules at init-time,
// keyed by their case-insensitive configuration name.
var registry = map[string]Mapper{}

// Register adds a named Mapper to the global registry.
//
// It is meant to be called from a plugin's init() function, so that importing
// the plugin package (typically as a blank/underscore import) makes the mapper
// available for resolution by name.
//
// Names are matched case-insensitively and are trimmed of surrounding whitespace.
// Registering an empty name, a nil Mapper, or a name that is already registered
// panics: a duplicate registration is a programming error that must be caught at
// startup rather than silently shadowing an existing mapper.
func Register(name string, m Mapper) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		panic("mapping: Register called with an empty name")
	}
	if m == nil {
		panic(fmt.Sprintf("mapping: Register called with a nil Mapper for name %q", name))
	}
	if _, exists := registry[key]; exists {
		panic(fmt.Sprintf("mapping: duplicate mapper registration for name %q", key))
	}
	registry[key] = m
}

// Resolve resolves the comma-separated mapper names in config against the registry.
//
// It returns the resolved mappers in configuration order, together with the list
// of unknown names: names that were configured but are not registered (typically
// because the plugin providing them was not imported). The caller is expected to
// log an explicit warning for any unknown name, so that a configured-but-unloaded
// mapper never silently disappears.
func Resolve(config string) (mappers []Mapper, unknown []string) {
	for _, raw := range strings.Split(config, ",") {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}

		if m, ok := registry[strings.ToLower(name)]; ok {
			mappers = append(mappers, m)
		} else {
			unknown = append(unknown, name)
		}
	}

	return mappers, unknown
}

// MappingsFromConfig detects supported mapping identifiers from the given config value,
// and returns a corresponding list of mapping functions resolved against the registry.
//
// Deprecated: use Resolve instead. MappingsFromConfig silently drops any configured
// name that is not registered, which hides a misconfigured or non-imported plugin.
// Resolve returns those unknown names so the caller can warn about them explicitly.
func MappingsFromConfig(config string) []Mapper {
	mappers, _ := Resolve(config)
	return mappers
}
