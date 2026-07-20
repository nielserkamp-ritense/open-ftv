FROM golang:1.26.5-alpine AS golang_builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/openftv ./external/gateways/envoy/plugins/openftv/cmd/

# Stage 2 - run
FROM envoyproxy/envoy:v1.31-latest

RUN apt-get update && apt-get install -y gettext && apt-get clean

COPY --from=golang_builder /build/openftv /usr/local/bin
RUN chmod 500 /usr/local/bin/openftv

COPY ./external/gateways/envoy/envoy/envoy.yaml /tmpl/envoy.yaml.tmpl
COPY ./external/gateways/envoy/envoy/envoy-entry.sh /
RUN chmod 500 /envoy-entry.sh

COPY ./external/gateways/envoy/docker-entrypoint.sh /
RUN chmod 500 /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
