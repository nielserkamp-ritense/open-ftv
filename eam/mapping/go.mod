module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping

go 1.23.8

toolchain go1.24.2

require (
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250415141202-eea5915a6d13
)

require (
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.17.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.etcd.io/etcd/client/v3 v3.5.21 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/oauth2 v0.29.0 // indirect
	golang.org/x/time v0.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/utils v0.0.0-20250321185631-1f6e0b77f77e // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mimetype => ../mimetype
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models => ../models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep => ../pep
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip => ../pip
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities => ../../utilities
)
