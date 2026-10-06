package pip

import (
	"io"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

// fileType represents the detected type of file.
type fileType uint8

// List of file types that can be detected.
const (
	unknownFile    fileType = iota // the type of file could not be detected.
	attributesFile                 // file contains attributes.
	entitiesFile                   // file contains entities.
	relationsFile                  // file contains relations.
	collectionFile                 // file contains a combination of attributes, entities and/or relations.
	policyFile                     // file contains a policy.
)

// detectType attempts to detect the type of file from the given stream and mime-type.
//
// If the given mime-type is empty, the function will attempt to detect the mime-type from the given stream.
// The return parameter *mimeType* may contain a different value again;
// e.g. when the mem-type appears to be 'application/json', but clearly contains JSON-LD content,
// the output parameter *mimeType* will be set to 'application/ld+json'.
//
// For supported mime-types, the function will detect what type of content the given stream contains,
// and return this value in the output parameter *ft*.
//
// When the content of the file needs to be examined, and if the type of content can then be determined,
// the output parameter *data* may contain the decoded content.
//
// In all other cases, it will return unknownFile for ft, mimeType will be empty and data will be nil.
func detectType(f io.ReadSeeker, mt string) (ft fileType, mimeType string, data any) {
	if mt == "" {
		mt = mime.SnifStream(f)
		_, _ = f.Seek(0, io.SeekStart)
	}

	switch mt {
	case mime.MimeTypeCedar, mime.MimeTypeCerbos, mime.MimeTypeOPA, mime.MimeTypeOpenFGA, mime.MimeTypeODRL, mime.MimeTypeXACML:
		return policyFile, mt, nil
	case mime.MimeTypeJSONLD, mime.MimeTypeNotation3, mime.MimeTypeRDF, mime.MimeTypeTurtle:
		return collectionFile, mt, nil
	case mime.MimeTypeJSON:
		return detectTypeFromJSON(f)
	case mime.MimeTypeYAML:
		return detectTypeFromYAML(f)
	case mime.MimeTypeTOML:
		return detectTypeFromTOML(f)
	}

	return
}

func detectTypeFromJSON(f io.ReadSeeker) (ft fileType, mimeType string, content any) {
	defer func() { _, _ = f.Seek(0, io.SeekStart) }()

	var data any
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return unknownFile, "", nil
	}

	switch t := data.(type) {
	case []any:
		return detectTypeFromSlice(t, mime.MimeTypeJSON)
	case map[string]any:
		if containsJSONLD(t) {
			return collectionFile, mime.MimeTypeJSONLD, data
		}

		return detectTypeFromMap(t, mime.MimeTypeJSON)
	default:
		return unknownFile, "", nil
	}
}

func detectTypeFromYAML(f io.ReadSeeker) (ft fileType, mimeType string, content any) {
	defer func() { _, _ = f.Seek(0, io.SeekStart) }()

	var data any
	if err := yaml.NewDecoder(f).Decode(&data); err != nil {
		return unknownFile, "", nil
	}

	switch t := data.(type) {
	case []any:
		return detectTypeFromSlice(t, mime.MimeTypeYAML)
	case map[string]any:
		return detectTypeFromMap(t, mime.MimeTypeYAML)
	default:
		return unknownFile, "", nil
	}
}

func detectTypeFromTOML(f io.ReadSeeker) (ft fileType, mimeType string, content any) {
	defer func() { _, _ = f.Seek(0, io.SeekStart) }()

	var data map[string]any
	if err := toml.NewDecoder(f).Decode(&data); err != nil {
		return unknownFile, "", nil
	}

	if len(data) == 0 {
		return attributesFile, mime.MimeTypeTOML, nil
	}

	return detectTypeFromMap(data, mime.MimeTypeTOML)
}

func detectTypeFromSlice(data []any, mt string) (ft fileType, mimeType string, content any) {
	for i := range data {
		if t, ok := data[i].(map[string]any); ok {
			if ft, _, _ := detectTypeFromMap(t, mt); ft != unknownFile {
				return ft, mt, data
			}
		}
	}

	return unknownFile, "", nil
}

func detectTypeFromMap(data map[string]any, mt string) (ft fileType, mimeType string, content any) {
	switch {
	case containsAttribute(data):
		return attributesFile, mt, data
	case containsEntity(data):
		return entitiesFile, mt, data
	case containsRelation(data):
		return relationsFile, mt, data
	default:
		return attributesFile, mt, data
	}
}
