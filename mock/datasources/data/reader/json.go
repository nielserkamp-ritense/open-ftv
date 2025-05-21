// Package reader contains functionality to read data-space and/or -source definitions and the actual data.
package reader

import (
	"io"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data/schema"
)

// DatasourceFromJSON loads a datasource definition from a JSON encoded data stream.
func DatasourceFromJSON(f io.Reader) (*schema.Datasource, error) {
	d := new(schema.Datasource)
	if err := json.NewDecoder(f).Decode(d); err != nil {
		return nil, err
	}
	return d, nil
}

// DataspaceFromJSON loads a dataspace definition from a JSON encoded data stream.
func DataspaceFromJSON(f io.Reader) (*schema.Dataspace, error) {
	d := new(schema.Dataspace)
	if err := json.NewDecoder(f).Decode(d); err != nil {
		return nil, err
	}
	return d, nil
}
