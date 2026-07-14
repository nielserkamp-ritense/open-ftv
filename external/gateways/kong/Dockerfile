FROM golang:1.26.5-alpine AS golang_builder

WORKDIR /build

COPY ./external/gateways/kong/plugins/openftv ./external/gateways/kong/plugins/openftv
COPY ./oas ./oas
COPY ./eam ./eam
COPY ./utilities ./utilities
COPY ./utilities-no-ci ./utilities-no-ci

RUN  cd external/gateways/kong/plugins/openftv \
  && go mod tidy \
  && go mod download \
  && CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/openftv ./cmd/*.go

# Stage 2 - run
FROM kong:3.9.1

COPY --from=golang_builder /build/openftv /usr/local/bin
