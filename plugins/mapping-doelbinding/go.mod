module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/plugins/mapping-doelbinding

go 1.23.8

toolchain go1.24.2

require (
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250708120440-2327e161c67a
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.17.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping => ../../eam/mapping
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mimetype => ../../eam/mimetype
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models => ../../eam/models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep => ../../eam/pep
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip => ../../eam/pip
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities => ../../utilities
)
