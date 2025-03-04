# Stage 1 - build
FROM digilabpublic.azurecr.io/golang:1.23.6-alpine3.21 AS golang_builder

WORKDIR /build

COPY ./go.* ./
COPY ./cmd/*.go ./

RUN go get gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas \
 && go get gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities \
 && go get gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci \
 && go get gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models \
 && go get gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components \
 && go get gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers \
 && go get gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server \
 && go get gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/apps/fsc/plugin/generic \
 && go mod tidy \
 && go mod download

COPY .. ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o fsc-auth ./*.go

# Stage 2 - run
FROM digilabpublic.azurecr.io/alpine:3.21

COPY --from=golang_builder /build/fsc-auth /

# Run the binary when starting the container
ENTRYPOINT ["/fsc-auth"]
EXPOSE 8443
