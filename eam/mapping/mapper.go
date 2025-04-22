package mapping

import (
	"strings"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
)

// Mapper is the function signature for mapping specific attributes from known locations into a preferred location.
//
// E,g, it can be used to map a specific attribute value as the principal, action or resource.
// This is useful to simplify policy selection based on attributes instead of the supplied principal, action and resource triplet.
type Mapper func(parc *models.PARC, opts ...Option) *models.PARC

// MappingsFromConfig detects supported mapping identifiers from the given config value,
// and returns a corresponding list of mapping functions.
func MappingsFromConfig(config string) []Mapper {
	var mappers []Mapper

	list := strings.Split(config, ",")
	for i := range list {
		switch strings.ToLower(strings.TrimSpace(list[i])) {
		case "rvvatoprincipal":
			mappers = append(mappers, RvvaToPrincipal)
		case "doelbindingtoprincipal":
			mappers = append(mappers, DoelbindingToPrincipal)
		case "bodytocontext":
			mappers = append(mappers, BodyToContext)
		}
	}

	return mappers
}

func (b *base) configure(opts []Option) {
	for i := range opts {
		opts[i](b)
	}
}

type base struct {
	headerKeys []string
}
