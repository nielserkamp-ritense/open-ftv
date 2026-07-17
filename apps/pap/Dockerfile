# Stage 1 - build
FROM golang:1.26.5-alpine AS golang_builder

WORKDIR /build

COPY ./apps/pap ./apps/pap
COPY ./oas ./oas
COPY ./eam ./eam
COPY ./utilities ./utilities
COPY ./migrations ./migrations
COPY ./utilities-no-ci ./utilities-no-ci

RUN cd apps/pap \
  && go mod tidy \
  && go mod download \
  && CGO_ENABLED=0 go build -ldflags="-s -w -X 'gitlab.com/digilab.overheid.nl/ftv/open-ftv/apps/pap/config/config.GitTag=${GIT_TAG_NAME}' -X 'gitlab.com/digilab.overheid.nl/ftv/open-ftv/apps/pap/config/config.GitHash=${GIT_COMMIT_HASH}' " -o /build/pap ./cmd/*.go

# Stage 2 - run
FROM alpine:3.22

RUN apk add --no-cache curl

COPY --from=golang_builder /build/pap /

# Run the binary when starting the container
ENTRYPOINT ["/pap"]
EXPOSE 8443
