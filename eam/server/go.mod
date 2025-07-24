module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server

go 1.24.4

require (
	github.com/goccy/go-json v0.10.5
	github.com/gofiber/fiber/v2 v2.52.9
	github.com/stretchr/testify v1.10.0
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas v0.0.0-20250708120440-2327e161c67a
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250708120440-2327e161c67a
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
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
	github.com/valyala/fasthttp v1.63.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas => ../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities => ../../utilities
)
