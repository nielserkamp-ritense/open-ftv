# Stage 1 - build
FROM digilabpublic.azurecr.io/golang:1.23.6-alpine3.21 AS golang_builder

WORKDIR /build

RUN mkdir -p /scripts

COPY ./remove_replace.sh /scripts/
COPY ./go.* ./
COPY ./cmd/* ./cmd/
COPY ./config/* ./config/
COPY ./persistence/* ./persistence/
COPY ./server/* ./server/

RUN chmod +x /scripts/remove_replace.sh
RUN /scripts/remove_replace.sh go.mod | tee go2.mod
RUN rm go.mod && mv go2.mod go.mod

RUN go get \
    gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas \
    gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities \
    gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities-no-ci \
    gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models \
    gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components \
    gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/handlers \
    gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server

RUN go mod tidy
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o pap ./cmd/*.go

# Stage 2 - run
FROM digilabpublic.azurecr.io/alpine:3.21

COPY --from=golang_builder /build/pap /

# Run the binary when starting the container
ENTRYPOINT ["/pap"]
EXPOSE 8443
