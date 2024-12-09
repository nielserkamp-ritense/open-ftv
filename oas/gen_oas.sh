#!/bin/bash

##### AuthZEN #####
cd authzen
oapi-codegen -config cfg-authzen.yaml oas-authzen.yaml
cd ..

##### FSC #####
# clone repo
cd fsc
git clone --depth 1 https://gitlab.com/commonground/nlx/fsc-nlx.git

cp fsc-nlx/outway/authorization-interface.yaml auth/oas-auth.yaml

# generate
cd auth
oapi-codegen -config cfg-auth.yaml oas-auth.yaml
cd ..

# remove repo
rm -Rf fsc-nlx
cd ..

##### FDS #####
# clone repo
cd fds
git clone --depth 1 https://gitlab.com/digilab.overheid.nl/ecosystem/fdsdemo.git

cp fdsdemo/fds/ledenlijst/api/openapi.yaml ledenlijst/oas-leden.yaml

# generate
cd ledenlijst
oapi-codegen -config cfg-leden.yaml oas-leden.yaml
sed -i '6,8d' ledenlijst.go
sed -i -e "s/openapi_types.UUID/string      /" ledenlijst.go
cd ..

# remove repo
rm -Rf fdsdemo
cd ..
