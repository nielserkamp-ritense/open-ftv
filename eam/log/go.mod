module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log

go 1.24.6

require (
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models v0.0.0-20250415141202-eea5915a6d13
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci v0.0.0-20250415141202-eea5915a6d13
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/defensestation/osquery v1.0.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/opensearch-project/opensearch-go v1.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas v0.0.0-20250908141655-a053f2619bdb // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250908141655-a053f2619bdb // indirect
	golang.org/x/exp v0.0.0-20250718183923-645b1fa84792 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models => ../models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci => ../../utilities-no-ci
)
