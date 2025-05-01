#!/bin/sh

VERSION=0.43.0
USER="cerbos"
PSWD="$(echo ${USER} | htpasswd -niBC 10 cerbos | cut -d ':' -f 2 | base64 -w0)"

docker run --rm --name cerbos-unittest -d \
  -v $(pwd)/../../testdata/policies/cerbos:/policies \
  -p 6693:3593 \
  ghcr.io/cerbos/cerbos:${VERSION} \
  server \
  --set=storage.driver=sqlite3 \
  --set=storage.sqlite3.dsn=file::memory:?cache=shared \
  --set=server.adminAPI.enabled=true \
  --set=server.adminAPI.adminCredentials.username=${USER} \
  --set=server.adminAPI.adminCredentials.passwordHash=${PSWD}
