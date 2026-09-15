# syntax=docker/dockerfile:1

# Stage 1 - build
FROM --platform=$BUILDPLATFORM golang:1.26.7-alpine AS golang_builder

ARG TARGETOS
ARG TARGETARCH
ARG GIT_TAG_NAME
ARG GIT_COMMIT_HASH

WORKDIR /build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w -X 'gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config.GitTag=${GIT_TAG_NAME}' -X 'gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config.GitHash=${GIT_COMMIT_HASH}'" -o /build/pdp ./apps/pdp/cmd/

# Stage 2 - run
FROM alpine:3.24

COPY --from=golang_builder /build/pdp /

# Run the binary when starting the container
ENTRYPOINT ["/pdp"]
EXPOSE 8443
