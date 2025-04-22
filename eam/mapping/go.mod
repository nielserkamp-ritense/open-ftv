module gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/mapping

go 1.23.2

require (
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pep v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities v0.0.0-20250415141202-eea5915a6d13
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.17.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250414145226-207652e42e2e // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250414145226-207652e42e2e // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/mimetype => ../mimetype
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models => ../models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pep => ../pep
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/pip => ../pip
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities => ../../utilities
)
