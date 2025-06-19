module gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server

go 1.23.8

toolchain go1.24.2

require (
	github.com/goccy/go-json v0.10.5
	github.com/gofiber/fiber/v2 v2.52.6
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas v0.0.0-20250410082626-e73c066519bd
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities v0.0.0-20250410082626-e73c066519bd
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.60.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas => ../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities => ../../utilities
)
