# Stage 1 - build
FROM golang:1.26.5-alpine AS golang_builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /build/generic-ds ./mock/datasources/generic/cmd/

# Stage 2 - run
FROM alpine:3.22

COPY --from=golang_builder /build/generic-ds /
COPY ./testdata/apps/gemeente-vlierdam/dataspace /opt/dataspace

ENV GEN_DS_DATA_PATH=/opt/dataspace

# Run the binary when starting the container
ENTRYPOINT ["/generic-ds"]
EXPOSE 8443
