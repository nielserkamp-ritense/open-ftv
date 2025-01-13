#!/bin/bash

##### AuthZEN #####
cd authzen
oapi-codegen -config cfg-authzen.yaml oas-authzen.yaml
sed -i -e "s/*string/string/" authzen.go
sed -i -e "s/*map/map/" authzen.go
sed -i -e "s/*\[\]/\[\]/" authzen.go
sed -i -e "s/*IgnoreMissing/IgnoreMissing/" authzen.go
sed -i -e "s/*ForceUpsert/ForceUpsert/" authzen.go
sed -i -e "s/*ReasonField/ReasonField/" authzen.go
cd ..

##### Policies #####
cd policies
oapi-codegen -config cfg-policies.yaml oas-policies.yaml
sed -i -e "s/*string/string/" policies.go
sed -i -e "s/*map/map/" policies.go
sed -i -e "s/*\[\]/\[\]/" policies.go
sed -i -e "s/*IgnoreMissing/IgnoreMissing/" policies.go
sed -i -e "s/*ForceUpsert/ForceUpsert/" policies.go
cd ..

##### Attributes #####
cd attributes
oapi-codegen -config cfg-attributes.yaml oas-attributes.yaml
sed -i -e "s/*string/string/" attributes.go
sed -i -e "s/*map/map/" attributes.go
sed -i -e "s/*\[\]/\[\]/" attributes.go
sed -i -e "s/*IgnoreMissing/IgnoreMissing/" attributes.go
sed -i -e "s/*ForceUpsert/ForceUpsert/" attributes.go
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
