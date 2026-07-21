# Stage 1 - build
FROM golang:1.26.5-alpine AS golang_builder

ARG GIT_TAG_NAME
ARG GIT_COMMIT_HASH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w -X 'gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config.GitTag=${GIT_TAG_NAME}' -X 'gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/apps/pdp/config.GitHash=${GIT_COMMIT_HASH}'" -o /build/pdp ./apps/pdp/cmd/

# Stage 2 - run
FROM alpine:3.22

RUN apk add --no-cache curl

COPY --from=golang_builder /build/pdp /

# Run the binary when starting the container
ENTRYPOINT ["/pdp"]
EXPOSE 8443
