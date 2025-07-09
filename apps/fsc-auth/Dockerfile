# Stage 1 - build
FROM digilabpublic.azurecr.io/golang:1.24.4-alpine AS golang_builder

WORKDIR /build

COPY ./apps/fsc-auth ./apps/fsc-auth
COPY ./oas ./oas
COPY ./eam ./eam
COPY ./utilities ./utilities
COPY ./utilities-no-ci ./utilities-no-ci

RUN cd apps/fsc-auth \
  && go mod tidy \
  && go mod download \
  && CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/fsc-auth ./cmd/*.go

# Stage 2 - run
FROM digilabpublic.azurecr.io/alpine:3.22

COPY --from=golang_builder /build/fsc-auth /

# Run the binary when starting the container
ENTRYPOINT ["/fsc-auth"]
EXPOSE 8443
