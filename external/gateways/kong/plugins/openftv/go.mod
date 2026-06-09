module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/external/gateways/kong/plugins/openftv

go 1.24.6

require (
	github.com/Kong/go-pdk v0.11.2
	github.com/MicahParks/keyfunc/v3 v3.8.0
	github.com/goccy/go-json v0.10.5
	github.com/golang-jwt/jwt/v5 v5.2.3
	github.com/stretchr/testify v1.11.1
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models v0.0.0-20250822123250-01bbd01a4df3
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep v0.0.0-20250822123250-01bbd01a4df3
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas v0.0.0-20251215100820-bfcd6fc0bc7c
)

require (
	github.com/MicahParks/jwkset v0.11.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20251215100820-bfcd6fc0bc7c // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/time v0.12.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models => ../../../../../eam/models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep => ../../../../../eam/pep
)
