# Stage 1 - build
FROM digilabpublic.azurecr.io/golang:1.24.6-alpine AS golang_builder

WORKDIR /build

COPY ./apps/pdp ./apps/pdp
COPY ./oas ./oas
COPY ./eam ./eam
COPY ./migrations ./migrations
COPY ./utilities ./utilities
COPY ./utilities-no-ci ./utilities-no-ci

RUN cd apps/pdp \
  && go mod tidy \
  && go mod download \
  && CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/pdp ./cmd/*.go

# Stage 2 - run
FROM digilabpublic.azurecr.io/alpine:3.22

COPY --from=golang_builder /build/pdp /

# Run the binary when starting the container
ENTRYPOINT ["/pdp"]
EXPOSE 8443
