module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/inzicht

go 1.23.8

toolchain go1.24.2

require (
	github.com/kvtools/valkeyrie v1.0.0
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config v0.0.0-20250708120440-2327e161c67a
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250708120440-2327e161c67a
	gitlab.com/gjuyn/go-config v1.2.0
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/config => ./../../eam/config
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log => ./../../eam/log
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mapping => ./../../eam/mapping
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/mimetype => ./../../eam/mimetype
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models => ./../../eam/models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas => ./../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities => ./../../utilities
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci => ./../../utilities-no-ci
)
