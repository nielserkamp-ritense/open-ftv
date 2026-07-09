module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl-geo

go 1.23.8

toolchain go1.24.2

require (
	github.com/deiu/rdf2go v0.0.0-20241212211204-b661ba0dfd25
	github.com/goccy/go-json v0.10.5
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models v0.0.0-20250415141202-eea5915a6d13
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250415141202-eea5915a6d13
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deiu/gon3 v0.0.0-20241212124032-93153c038193 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/kvtools/valkeyrie v1.0.0 // indirect
	github.com/linkeddata/gojsonld v0.0.0-20170418210642-4f5db6791326 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rychipman/easylex v0.0.0-20160129204217-49ee7767142f // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping => ../../mapping
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models => ../../models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap => ../../pap
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller => ../controller
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl => ../odrl
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pep => ../../pep
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip => ../../pip
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas => ../../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities => ../../../utilities
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci => ../../../utilities-no-ci
)
