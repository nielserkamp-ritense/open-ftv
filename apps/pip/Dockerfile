# Stage 1 - build
FROM digilabpublic.azurecr.io/golang:1.23.8-alpine AS golang_builder

WORKDIR /build

COPY ./apps/pip ./apps/pip
COPY ./oas ./oas
COPY ./eam ./eam
COPY ./utilities ./utilities
COPY ./utilities-no-ci ./utilities-no-ci

RUN cd apps/pip \
  && go mod tidy \
  && go mod download \
  && CGO_ENABLED=0 go build -ldflags="-s -w" -o /build/pip ./cmd/*.go

# Stage 2 - run
FROM digilabpublic.azurecr.io/alpine:3.21

COPY --from=golang_builder /build/pip /

# Run the binary when starting the container
ENTRYPOINT ["/pip"]
EXPOSE 8443
