package models

// Meta defines the paths for loading definitions and data from disk storage.
type Meta struct {
	DataspaceDef   string                       `yaml:"dataspace,omitempty"`
	DatasourceDefs []string                     `yaml:"datasource,omitempty"`
	EndpointDefs   []string                     `yaml:"endpoints,omitempty"`
	SourceData     map[string]map[string]string `yaml:"sourceData,omitempty"`
}
