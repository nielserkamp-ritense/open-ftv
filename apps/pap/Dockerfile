# Stage 1 - build
FROM digilabpublic.azurecr.io/golang:1.22.8-alpine3.19 AS golang_builder

WORKDIR /build

COPY ./go.* ./
COPY ./cmd/*.go ./

RUN go mod download

COPY .. ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o pap ./*.go

# Stage 2 - run
FROM digilabpublic.azurecr.io/alpine:3.19

COPY --from=golang_builder /build/pap /

# Run the binary when starting the container
ENTRYPOINT ["/pap"]
EXPOSE 8443
