package pip

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mimetype"
	mime2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

func (p *pip) loadAttributes(path string) {
	f, err := os.Open(path)
	if err != nil {
		p.logger.Error("pip: error opening attributes file", "path", path, "err", err)
		return
	}
	defer f.Close()

	// A file-store load registers a "file" source reference for every loaded value.
	ref := p.fileSourceRef(path)

	// detect the file-type and mime-type if possible.
	// this should cover all supported file types.
	if ft, mt, data := mime.DetectType(f, mime2.ConvertExt(filepath.Ext(path))); ft == mime.AttributesFile || ft == mime.CollectionFile {
		if data != nil {
			// this covers YAML, TOML & JSON.
			p.loadAttributesAny(data, ref)
			return
		}

		switch mt {
		case mime2.MimeTypeTurtle, mime2.MimeTypeJSONLD, mime2.MimeTypeRDF:
			// this covers our supported RDF file encodings.
			p.loadRDF(f, path, mt, ref)
			return
		}
	}

	// detection failed, so we'll assume YAML and hope for the best.
	var attributes any
	if err = yaml.NewDecoder(f).Decode(&attributes); err != nil {
		p.logger.Error("pip: error decoding attributes file (YAML)", "path", path, "err", err)
		return
	}

	p.loadAttributesAny(attributes, ref)
}

func (p *pip) loadAttributesAny(attributes any, ref *SourceRef) {
	switch t := attributes.(type) {
	case []any:
		for i := range t {
			p.loadAttributesAny(t[i], ref)
		}
	case []map[string]any:
		for i := range t {
			p.loadAttributeMap(t[i], ref)
		}
	case map[string]any:
		p.loadAttributeMap(t, ref)
	}
}

func (p *pip) loadAttributeMap(attribute map[string]any, ref *SourceRef) {
	k, ok1 := attribute["key"].(string)
	v, ok2 := attribute["value"]
	if ok1 && ok2 {
		p.AddAttribute(k, v)
		p.recordFileRef(k, ref)
	}
}
