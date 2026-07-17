# Stage 1 - build
FROM golang:1.26.5-alpine AS golang_builder

WORKDIR /build

COPY ./apps/manager ./apps/manager
COPY ./oas ./oas
COPY ./eam ./eam
COPY ./migrations ./migrations
COPY ./utilities ./utilities
COPY ./utilities-no-ci ./utilities-no-ci

RUN cd apps/manager \
  && go mod tidy \
  && go mod download \
  && CGO_ENABLED=0 go build -ldflags="-s -w -X 'gitlab.com/digilab.overheid.nl/ftv/open-ftv/apps/manager/config/config.GitTag=${GIT_TAG_NAME}' -X 'gitlab.com/digilab.overheid.nl/ftv/open-ftv/apps/manager/config/config.GitHash=${GIT_COMMIT_HASH}' " -o /build/manager ./cmd/*.go

# Stage 2 - run
FROM alpine:3.22

RUN apk add --no-cache curl

COPY --from=golang_builder /build/manager /

# Run the binary when starting the container
ENTRYPOINT ["/manager"]
EXPOSE 8443
