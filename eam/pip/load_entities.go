package pip

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mimetype"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	mime2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

func (p *pip) loadEntities(path string) {
	f, err := os.Open(path)
	if err != nil {
		p.logger.Error("pip: error opening entities file", "path", path, "err", err)
		return
	}
	defer f.Close()

	// A file-store load registers a "file" source reference for every loaded value,
	// so the ADL level-3 hook can resolve where a file-backed attribute came from.
	ref := p.fileSourceRef(path)

	// detect the file-type and mime-type if possible.
	// this should cover all supported file types.
	if ft, mt, data := mime.DetectType(f, mime2.ConvertExt(filepath.Ext(path))); ft == mime.EntitiesFile || ft == mime.CollectionFile {
		if data != nil {
			// this covers YAML, TOML & JSON.
			p.loadEntitiesAny(data, ref)
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
	var entities any
	if err = yaml.NewDecoder(f).Decode(&entities); err != nil {
		p.logger.Error("pip: error decoding entities file (YAML)", "path", path, "err", err)
		return
	}

	p.loadEntitiesAny(entities, ref)
}

func (p *pip) loadEntitiesAny(entities any, ref *SourceRef) {
	switch t := entities.(type) {
	case []any:
		for i := range t {
			p.loadEntitiesAny(t[i], ref)
		}
	case []map[string]any:
		for i := range t {
			p.loadEntityMap(t[i], ref)
		}
	case map[string]any:
		p.loadEntityMap(t, ref)
	}
}

func (p *pip) loadEntityMap(attribute map[string]any, ref *SourceRef) {
	t, ok1 := attribute["type"].(string)
	id, ok2 := attribute["id"].(string)

	if ok1 && ok2 {
		attrs := p.newAttributes()
		if q := attribute["attributes"]; q != nil {
			if m, ok3 := q.(map[string]any); ok3 {
				for k := range m {
					attrs.AddAttribute(k, m[k])
				}
			}
		}

		var parents []string
		if q := attribute["parents"]; q != nil {
			if q2, ok3 := q.([]any); ok3 {
				for i := range q2 {
					if s, ok := q2[i].(string); ok {
						parents = append(parents, s)
					}
				}
			}
		}

		e := models.NewEntity(t, id, attrs, parents...)
		p.AddEntity(e)
		p.recordEntityFileRef(e, ref)
	}
}
