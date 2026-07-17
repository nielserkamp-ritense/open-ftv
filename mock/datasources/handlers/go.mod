module gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/handlers

go 1.26.5

require (
	github.com/goccy/go-json v0.10.5
	github.com/goccy/go-yaml v1.18.0
	github.com/gofiber/fiber/v2 v2.52.13
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server v0.0.0-20250708120440-2327e161c67a
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data v0.0.0-00010101000000-000000000000
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities v0.0.0-20250708120440-2327e161c67a
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.63.0 // indirect
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas v0.0.0-20250708120440-2327e161c67a // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/sys v0.33.0 // indirect
)

replace (
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server => ../../../eam/server
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/mock/datasources/data => ../data
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas => ../../../oas
	gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities => ../../../utilities
)
