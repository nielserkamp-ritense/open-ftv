// Package mimetype contains functionality for determining types of data.
package mimetype

import (
	"io"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

// FileType represents the detected type of file.
type FileType uint8

// List of file types that can be detected.
const (
	UnknownFile    FileType = iota // the type of file could not be detected.
	AttributesFile                 // file contains attributes.
	EntitiesFile                   // file contains entities.
	RelationsFile                  // file contains relations.
	CollectionFile                 // file contains a combination of attributes, entities and/or relations.
	PolicyFile                     // file contains a policy.
)

// DetectType attempts to detect the type of file from the given stream and mime-type.
//
// If the given mime-type is empty, the function will attempt to detect the mime-type from the given stream.
// The return parameter *mimeType* may contain a different value again;
// e.g. when the mem-type appears to be 'application/json', but clearly contains JSON-LD content,
// the output parameter *mimeType* will be set to 'application/ld+json'.
//
// For supported mime-types, the function will detect what type of content the given stream contains,
// and return this value in the output parameter *fileType*.
//
// When the content of the file needs to be examined, and if the type of content can then be determined,
// the output parameter *data* may contain the decoded content.
//
// In all other cases, it will return UnknownFile for fileType, mimeType will be empty and data will be nil.
func DetectType(f io.ReadSeeker, mt string) (fileType FileType, mimeType string, data any) {
	if mt == "" {
		mt = mime.SnifStream(f)
		_, _ = f.Seek(0, io.SeekStart)
	}

	switch mt {
	case mime.MimeTypeCedar, mime.MimeTypeCerbos, mime.MimeTypeOPA, mime.MimeTypeOpenFGA, mime.MimeTypeODRL, mime.MimeTypeXACML:
		return PolicyFile, mt, nil
	case mime.MimeTypeJSONLD, mime.MimeTypeNotation3, mime.MimeTypeRDF, mime.MimeTypeTurtle:
		return CollectionFile, mt, nil
	case mime.MimeTypeJSON:
		return detectTypeFromJSON(f)
	case mime.MimeTypeYAML:
		return detectTypeFromYAML(f)
	case mime.MimeTypeTOML:
		return detectTypeFromTOML(f)
		// case mime.MimeTypeXML:
		// 	// TODO: ...
		// case mime.MimeTypeCSV:
		// 	// TODO: ...
		// case mime.MimeTypeTabSep:
		// 	// TODO: ...
	}

	return
}

func detectTypeFromJSON(f io.ReadSeeker) (FileType, string, any) {
	defer func() { _, _ = f.Seek(0, io.SeekStart) }()

	var data any
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return UnknownFile, "", nil
	}

	switch t := data.(type) {
	case []any:
		return detectTypeFromSlice(t, mime.MimeTypeJSON)
	case map[string]any:
		if ContainsJSONLD(t) {
			return CollectionFile, mime.MimeTypeJSONLD, data
		}
		return detectTypeFromMap(t, mime.MimeTypeJSON)
	default:
		return UnknownFile, "", nil
	}
}

func detectTypeFromYAML(f io.ReadSeeker) (FileType, string, any) {
	defer func() { _, _ = f.Seek(0, io.SeekStart) }()

	var data any
	if err := yaml.NewDecoder(f).Decode(&data); err != nil {
		return UnknownFile, "", nil
	}

	switch t := data.(type) {
	case []any:
		return detectTypeFromSlice(t, mime.MimeTypeYAML)
	case map[string]any:
		return detectTypeFromMap(t, mime.MimeTypeYAML)
	default:
		return UnknownFile, "", nil
	}
}

func detectTypeFromTOML(f io.ReadSeeker) (FileType, string, any) {
	defer func() { _, _ = f.Seek(0, io.SeekStart) }()

	var data map[string]any
	if err := toml.NewDecoder(f).Decode(&data); err != nil {
		return UnknownFile, "", nil
	}
	if len(data) == 0 {
		return AttributesFile, mime.MimeTypeTOML, nil
	}
	return detectTypeFromMap(data, mime.MimeTypeTOML)
}

func detectTypeFromSlice(data []any, mt string) (FileType, string, any) {
	for i := range data {
		switch t := data[i].(type) {
		case map[string]any:
			if ft, _, _ := detectTypeFromMap(t, mt); ft != UnknownFile {
				return ft, mt, data
			}
		}
	}
	return UnknownFile, "", nil
}

func detectTypeFromMap(data map[string]any, mt string) (FileType, string, any) {
	switch {
	case ContainsAttribute(data):
		return AttributesFile, mt, data
	case ContainsEntity(data):
		return EntitiesFile, mt, data
	case ContainsRelation(data):
		return RelationsFile, mt, data
	default:
		return AttributesFile, mt, data
	}
}
