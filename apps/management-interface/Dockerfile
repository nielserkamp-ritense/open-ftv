# Stage 1 - build the static assets
FROM --platform=$BUILDPLATFORM docker.io/node:22.17.1-alpine AS assets_builder

WORKDIR /app

COPY ./apps/management-interface/package.json ./apps/management-interface/package-lock.json /app/
RUN npm ci

COPY ./apps/management-interface/ .
RUN npm run build

# Stage 2 - build the web server with the assets embedded
FROM --platform=$BUILDPLATFORM golang:1.26.7-alpine AS golang_builder

ARG TARGETOS
ARG TARGETARCH
ARG GIT_TAG_NAME
ARG GIT_COMMIT_HASH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=assets_builder /app/web/dist ./apps/management-interface/web/dist

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w -X 'gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/management-interface/config.GitTag=${GIT_TAG_NAME}' -X 'gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/management-interface/config.GitHash=${GIT_COMMIT_HASH}'" -o /build/management-interface ./apps/management-interface/cmd/

# Stage 3 - run
FROM alpine:3.22

COPY --from=golang_builder /build/management-interface /

# Run the binary when starting the container
ENTRYPOINT ["/management-interface"]
EXPOSE 8080
