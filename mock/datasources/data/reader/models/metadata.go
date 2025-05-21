package models

// Meta defines the paths for loading definitions and data from disk storage.
type Meta struct {
	DataspaceDef   string                       `yaml:"dataspace,omitempty"`
	DatasourceDefs []string                     `yaml:"datasource,omitempty"`
	SourceData     map[string]map[string]string `yaml:"sourceData,omitempty"`
}
