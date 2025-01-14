#!/bin/bash

function fix_code {
  sed -i -e "s/*string/string/" $1.go
  sed -i -e "s/*map/map/" $1.go
  sed -i -e "s/*\[\]/\[\]/" $1.go
  sed -i -e "s/*IgnoreMissing/IgnoreMissing/" $1.go
  sed -i -e "s/*ForceUpsert/ForceUpsert/" $1.go
  sed -i -e "s/*ReasonField/ReasonField/" $1.go
  sed -i -e "s/interface{}/any/" $1.go
}

##### AuthZEN #####
cd authzen
oapi-codegen -config cfg-authzen.yaml oas-authzen.yaml
fix_code "authzen"
cd ..

##### Policies #####
cd policies
oapi-codegen -config cfg-policies.yaml oas-policies.yaml
fix_code "policies"
cd ..

##### Attributes #####
cd attributes
oapi-codegen -config cfg-attributes.yaml oas-attributes.yaml
fix_code "attributes"
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
