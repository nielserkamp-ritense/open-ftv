module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci

go 1.26.5

require (
	github.com/defensestation/osquery v1.0.0
	github.com/goccy/go-json v0.10.5
	github.com/goccy/go-yaml v1.18.0
	github.com/golang-jwt/jwt/v5 v5.2.3
	github.com/google/go-github/v69 v69.2.0
	github.com/google/uuid v1.6.0
	github.com/jferrl/go-githubauth v1.2.0
	github.com/opensearch-project/opensearch-go v1.1.0
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250708120440-2327e161c67a
	golang.org/x/oauth2 v0.30.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/time v0.12.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities => ./../utilities
