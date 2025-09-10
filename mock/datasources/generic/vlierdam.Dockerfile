# Stage 1 - build
FROM digilabpublic.azurecr.io/golang:1.24.6-alpine AS golang_builder

WORKDIR /build

COPY ./oas ./oas
COPY ./migrations ./migrations
COPY ./utilities ./utilities
COPY ./utilities-no-ci ./utilities-no-ci
COPY ./eam ./eam
COPY ./mock/datasources ./mock/datasources
COPY ./testdata/apps/gemeente-vlierdam/dataspace ./opt/data

RUN cd mock/datasources/generic \
  && go mod tidy \
  && go mod download \
  && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/generic-ds ./cmd/*.go

# Stage 2 - run
FROM digilabpublic.azurecr.io/alpine:3.22

COPY --from=golang_builder /build/generic-ds /
COPY ./testdata/apps/gemeente-vlierdam/dataspace ./opt/dataspace

ENV GEN_DS_DATA_PATH=/opt/dataspace

# Run the binary when starting the container
ENTRYPOINT ["/generic-ds"]
EXPOSE 8443
