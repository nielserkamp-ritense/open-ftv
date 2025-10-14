package config

// OpenSearch contains the configuration variables for connecting with an OpenSearch cluster.
type OpenSearch struct {
	Endpoints string `json:"osEndpoints,omitempty" yaml:"opensearch.endpoints,omitempty" env:"OPENSEARCH_ENDPOINTS" flag:"opensearch-endpoints" desc:"OpenSearch endpoints"`
	Index     string `json:"osIndex,omitempty"     yaml:"opensearch.index,omitempty"     env:"OPENSEARCH_INDEX"     flag:"opensearch-index"     desc:"OpenSearch index name"`
	User      string `json:"-"                     yaml:"opensearch.user,omitempty"      env:"OPENSEARCH_USER"      flag:"opensearch-user"      desc:"OpenSearch username"`
	Pswd      string `json:"-"                     yaml:"opensearch.password,omitempty"  env:"OPENSEARCH_PASSWORD"  flag:"opensearch-password"  desc:"OpenSearch password"`
}

// Sanitized returns the configuration variables where all sensitive data has been scrubbed.
func (o *OpenSearch) Sanitized() *OpenSearch {
	sanitized := *o
	sanitized.User = ""
	sanitized.Pswd = ""
	return &sanitized
}
