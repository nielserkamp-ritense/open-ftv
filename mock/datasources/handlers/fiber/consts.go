package fiber

// List of API paths.
const (
	PathDataspace    = "/dataspace"
	PathDatasources  = "/datasources"
	PathDatasource   = "/datasource/:key"
	PathTables       = "/tables"
	PathTable        = "/table/:key"
	PathTableRecords = "/table/:key/records"
	PathTableRecord  = "/table/:key/record"
)

// List of mime-types.
const (
	MimeYAML1 = "application/yaml"
	MimeYAML2 = "text/yaml"
	MimeJSON2 = "text/json"
	MimeCSV   = "text/csv"
)

// HeaderVersion is the header for reporting the full semantic API version.
const HeaderVersion = "API-Version"
