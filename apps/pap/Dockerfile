# Stage 1 - build
FROM digilabpublic.azurecr.io/golang:1.23.6-alpine3.21 AS golang_builder

WORKDIR /build

COPY ./apps/pap ./apps/pap
COPY ./oas ./oas
COPY ./eam ./eam
COPY ./utilities ./utilities
COPY ./utilities-no-ci ./utilities-no-ci

RUN cd apps/pap \
  && go mod tidy \
  && go mod download \
  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/pap ./cmd/*.go

# Stage 2 - run
FROM digilabpublic.azurecr.io/alpine:3.21

COPY --from=golang_builder /build/pap /

# Run the binary when starting the container
ENTRYPOINT ["/pap"]
EXPOSE 8443
