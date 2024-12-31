package pip

import (
	"io"

	"github.com/goccy/go-json"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/io"
)

type fileType uint8

const (
	unknownFile fileType = iota
	attributesFile
	entitiesFile
	relationsFile
	collectionFile
	policyFile
)

func detectType(f io.ReadSeeker, mt string) (fileType, string, any) {
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
	case mime.MimeTypeXML:
		// TODO: ...
	case mime.MimeTypeCSV:
		// TODO: ...
	case mime.MimeTypeTabSep:
		// TODO: ...
	}

	return unknownFile, "", nil
}

func detectTypeFromJSON(f io.Reader) (fileType, string, any) {
	var data any
	if err := json.NewDecoder(f).Decode(&data); err != nil {
		return unknownFile, "", nil
	}

	switch t := data.(type) {
	case []any:
		return detectTypeFromSlice(t, mime.MimeTypeJSON)
	case map[string]any:
		if isJsonLD(t) {
			return collectionFile, mime.MimeTypeJSONLD, data
		}
		return detectTypeFromMap(t, mime.MimeTypeJSON)
	default:
		return unknownFile, "", nil
	}
}

func detectTypeFromYAML(f io.Reader) (fileType, string, any) {
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

func detectTypeFromTOML(f io.Reader) (fileType, string, any) {
	var data any
	if err := toml.NewDecoder(f).Decode(&data); err != nil {
		return unknownFile, "", nil
	}

	switch t := data.(type) {
	case []any:
		return detectTypeFromSlice(t, mime.MimeTypeTOML)
	case map[string]any:
		return detectTypeFromMap(t, mime.MimeTypeTOML)
	default:
		return unknownFile, "", nil
	}
}

func detectTypeFromSlice(data []any, mt string) (fileType, string, any) {
	for i := range data {
		switch t := data[i].(type) {
		case map[string]any:
			if ft, _, _ := detectTypeFromMap(t, mt); ft != unknownFile {
				return ft, mt, data
			}
		}
	}
	return unknownFile, "", nil
}

func detectTypeFromMap(data map[string]any, mt string) (fileType, string, any) {
	switch {
	case data["key"] != nil && data["value"] != nil:
		return attributesFile, mt, data
	case data["type"] != nil && data["id"] != nil:
		return entitiesFile, mt, data
	case data["subject"] != nil && data["predicate"] != nil && data["object"] != nil:
		return relationsFile, mt, data
	default:
		return attributesFile, mt, data
	}
}

func isJsonLD(data map[string]any) bool {
	return data["@context"] != nil || data["@id"] != nil || data["@graph"] != nil
}
