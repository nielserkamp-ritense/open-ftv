module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/external/gateways/kong/plugins/openftv

go 1.24.6

require (
	github.com/Kong/go-pdk v0.11.2
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models v0.0.0-20250822123250-01bbd01a4df3
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep v0.0.0-20250822123250-01bbd01a4df3
)

require (
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas v0.0.0-20250825104223-beef20b40619 // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250825104223-beef20b40619 // indirect
	golang.org/x/exp v0.0.0-20250718183923-645b1fa84792 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models => ../../../../../eam/models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep => ../../../../../eam/pep
)
