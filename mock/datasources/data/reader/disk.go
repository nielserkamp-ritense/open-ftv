package reader

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/reader/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/store"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/io"
)

// LoadFromPath loads all data-space and -source definitions from disk,
// as well as any data found in the given path or its subdirectories.
func LoadFromPath(s store.Storage, basePath string) error {
	p := &pathLoader{
		basePath: path.Clean(basePath),
		store:    s,
	}
	return p.run()
}

func (p *pathLoader) run() error {
	ok, err := p.findMeta()
	if err != nil {
		return err
	}

	if ok {
		return p.processMeta()
	}
	return fmt.Errorf("no metadata found")
}

func (p *pathLoader) findMeta() (bool, error) {
	err := filepath.WalkDir(p.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if strings.EqualFold(filepath.Base(path), "metadata.yaml") || strings.ToLower(filepath.Ext(path)) == ".meta" {
			if err = p.loadMeta(path); err != nil {
				p.meta = nil
				return err
			}
			return filepath.SkipAll
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	return p.meta != nil, nil
}

func (p *pathLoader) loadMeta(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	p.meta = new(models.Meta)
	return yaml.NewDecoder(f).Decode(p.meta)
}

func (p *pathLoader) processMeta() error {
	if s := p.meta.DataspaceDef; s != "" {
		if err := p.loadDataspace(p.fixPath(s)); err != nil {
			return err
		}
	}

	for i := range p.meta.DatasourceDefs {
		s := p.meta.DatasourceDefs[i]
		if err := p.loadDatasource(p.fixPath(s)); err != nil {
			return err
		}
	}

	for source := range p.meta.SourceData {
		m := p.meta.SourceData[source]
		for table := range m {
			if err := p.loadTableData(source, table, p.fixPath(m[table])); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *pathLoader) fixPath(path string) string {
	if strings.HasPrefix(path, "/") {
		return path
	}
	return filepath.Join(p.basePath, path)
}

func (p *pathLoader) loadDataspace(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("dataspace [%s]; %w", path, err)
	}
	defer f.Close()

	var t string
	if t = io.ConvertExt(filepath.Ext(path)); !io.IsSupported(t) {
		t = io.SnifStream(f)
	}

	var def *schema.Dataspace
	switch t {
	case io.MimeTypeJSON:
		def, err = DataspaceFromJSON(f)
	case io.MimeTypeYAML:
		def, err = DataspaceFromYAML(f)
	default:
		err = fmt.Errorf("unsupported data format: %s", t)
	}
	if err != nil {
		return fmt.Errorf("dataspace [%s]; %w", path, err)
	}

	p.store.SetDataspace(def)
	return nil
}

func (p *pathLoader) loadDatasource(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("datasource [%s]; %w", path, err)
	}
	defer f.Close()

	var t string
	if t = io.ConvertExt(filepath.Ext(path)); !io.IsSupported(t) {
		t = io.SnifStream(f)
	}

	var def *schema.Datasource
	switch t {
	case io.MimeTypeJSON:
		def, err = DatasourceFromJSON(f)
	case io.MimeTypeYAML:
		def, err = DatasourceFromYAML(f)
	default:
		err = fmt.Errorf("unsupported data format: %s", t)
	}
	if err != nil {
		return fmt.Errorf("datasource [%s]; %w", path, err)
	}

	p.store.AddDatasource(def)
	return nil
}

func (p *pathLoader) loadTableData(source, table, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("table data [%s]; %w", path, err)
	}
	defer f.Close()

	var t string
	if t = io.ConvertExt(filepath.Ext(path)); !io.IsSupported(t) {
		t = io.SnifStream(f)
	}

	var records []map[string]any
	var csvRecords [][]string

	switch t {
	case io.MimeTypeJSON:
		err = json.NewDecoder(f).Decode(&records)
	case io.MimeTypeYAML:
		err = yaml.NewDecoder(f).Decode(&records)
	case io.MimeTypeCSV:
		csvRecords, err = csv.NewReader(f).ReadAll()
	default:
		err = fmt.Errorf("unsupported file type: %s", t)
	}
	if err != nil {
		return fmt.Errorf("table data [%s]; %w", path, err)
	}

	if csvRecords != nil {
		err = p.store.AddTableFromCSV(source, table, csvRecords)
	} else {
		err = p.store.AddTableFromData(source, table, records)
	}

	if err != nil {
		return fmt.Errorf("table data [%s]; %w", path, err)
	}
	return nil
}

type pathLoader struct {
	basePath string
	store    store.Storage
	meta     *models.Meta
}
