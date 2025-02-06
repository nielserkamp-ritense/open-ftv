module gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server

go 1.23.0

toolchain go1.23.2

require (
	github.com/goccy/go-json v0.10.5
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers v0.0.0-20250205102328-4cadc1a4635f
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities v0.0.0-20250205102328-4cadc1a4635f
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deiu/gon3 v0.0.0-20241212124032-93153c038193 // indirect
	github.com/deiu/rdf2go v0.0.0-20241212211204-b661ba0dfd25 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/goccy/go-yaml v1.15.17 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/linkeddata/gojsonld v0.0.0-20170418210642-4f5db6791326 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/opensearch-project/opensearch-go v1.1.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rychipman/easylex v0.0.0-20160129204217-49ee7767142f // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.58.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components v0.0.0-20250205102328-4cadc1a4635f // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models v0.0.0-20250205102328-4cadc1a4635f // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas v0.0.0-20250205102328-4cadc1a4635f // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci v0.0.0-00010101000000-000000000000 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components => ./../components
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers => ./../handlers
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models => ./../models
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities => ../../utilities
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci => ../../utilities-no-ci
)
