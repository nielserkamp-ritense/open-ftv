package pip

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/mimetype"
	mime2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/io"
)

func (p *pip) loadAttributes(path string) {
	f, err := os.Open(path)
	if err != nil {
		p.logger.Error("pip: error opening attributes file", "path", path, "err", err)
		return
	}
	defer f.Close()

	// detect the file-type and mime-type if possible.
	// this should cover all supported file types.
	if ft, mt, data := mime.DetectType(f, mime2.ConvertExt(filepath.Ext(path))); ft == mime.AttributesFile || ft == mime.CollectionFile {
		if data != nil {
			// this covers YAML, TOML & JSON.
			p.loadAttributesAny(data)
			return
		}

		switch mt {
		case mime2.MimeTypeTurtle, mime2.MimeTypeJSONLD, mime2.MimeTypeRDF:
			// this covers our supported RDF file encodings.
			p.loadRDF(f, path, mt)
			return
		}
	}

	// detection failed, so we'll assume YAML and hope for the best.
	var attributes any
	if err = yaml.NewDecoder(f).Decode(&attributes); err != nil {
		p.logger.Error("pip: error decoding attributes file (YAML)", "path", path, "err", err)
		return
	}

	p.loadAttributesAny(attributes)
}

func (p *pip) loadAttributesAny(attributes any) {
	switch t := attributes.(type) {
	case []any:
		for i := range t {
			p.loadAttributesAny(t[i])
		}
	case []map[string]any:
		for i := range t {
			p.loadAttributeMap(t[i])
		}
	case map[string]any:
		p.loadAttributeMap(t)
	}
}

func (p *pip) loadAttributeMap(attribute map[string]any) {
	k, ok1 := attribute["key"].(string)
	v, ok2 := attribute["value"]
	if ok1 && ok2 {
		p.AddAttribute(k, v)
	}
}
