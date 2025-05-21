module gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/mock/datasources/data

go 1.23.8

toolchain go1.24.2

require (
	github.com/goccy/go-json v0.10.5
	github.com/goccy/go-yaml v1.17.1
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities => ../../../utilities
