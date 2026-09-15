# Stage 1 - build
FROM --platform=$BUILDPLATFORM golang:1.26.7-alpine AS golang_builder

ARG TARGETOS
ARG TARGETARCH
ARG GIT_TAG_NAME
ARG GIT_COMMIT_HASH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w -X 'gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config.GitTag=${GIT_TAG_NAME}' -X 'gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/manager/config.GitHash=${GIT_COMMIT_HASH}'" -o /build/manager ./apps/manager/cmd/

# Stage 2 - run
FROM alpine:3.24

COPY --from=golang_builder /build/manager /

# Run the binary when starting the container
ENTRYPOINT ["/manager"]
EXPOSE 8443
