package config

// OpenSearch contains the configuration variables for connecting with an OpenSearch cluster.
type OpenSearch struct {
	Endpoints string `yaml:"opensearch.endpoints,omitempty" env:"OPENSEARCH_ENDPOINTS" flag:"opensearch-endpoints" desc:"OpenSearch endpoints"`
	Index     string `yaml:"opensearch.index,omitempty" env:"OPENSEARCH_INDEX" flag:"opensearch-index" desc:"OpenSearch index name"`
	User      string `yaml:"opensearch.user,omitempty" env:"OPENSEARCH_USER" flag:"opensearch-user" desc:"OpenSearch username"`
	Pswd      string `yaml:"opensearch.password,omitempty" env:"OPENSEARCH_PASSWORD" flag:"opensearch-password" desc:"OpenSearch password"`
}

// Sanitized returns the configuration variables where all sensitive data has been scrubbed.
func (o *OpenSearch) Sanitized() *OpenSearch {
	sanitized := *o
	sanitized.User = ""
	sanitized.Pswd = ""
	return &sanitized
}
