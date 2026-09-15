# Stage 1 - build
FROM --platform=$BUILDPLATFORM golang:1.26.7-alpine AS golang_builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /build/generic-ds ./mock/datasources/generic/cmd/

# Stage 2 - run
FROM alpine:3.24

COPY --from=golang_builder /build/generic-ds /

# Run the binary when starting the container
ENTRYPOINT ["/generic-ds"]
EXPOSE 8443
