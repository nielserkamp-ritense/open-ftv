package reader

import (
	"io"

	"github.com/goccy/go-yaml"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data/schema"
)

// DatasourceFromYAML loads a datasource definition from a YAML encoded data stream.
func DatasourceFromYAML(f io.Reader) (*schema.Datasource, error) {
	d := new(schema.Datasource)
	if err := yaml.NewDecoder(f).Decode(d); err != nil {
		return nil, err
	}
	return d, nil
}

// DataspaceFromYAML loads a dataspace definition from a YAML encoded data stream.
func DataspaceFromYAML(f io.Reader) (*schema.Dataspace, error) {
	d := new(schema.Dataspace)
	if err := yaml.NewDecoder(f).Decode(d); err != nil {
		return nil, err
	}
	return d, nil
}
