#!/bin/bash

function fix_code {
  sed -i -e "s/*string /string /" $1.go
  sed -i -e "s/*map/map/" $1.go
  sed -i -e "s/*int /int /" $1.go
  sed -i -e "s/*\[\]/\[\]/" $1.go
  sed -i -e "s/*IgnoreMissing/IgnoreMissing/" $1.go
  sed -i -e "s/*ForceUpsert/ForceUpsert/" $1.go
  sed -i -e "s/*ReasonField/ReasonField/" $1.go
  sed -i -e "s/interface{}/any/" $1.go
  go fmt $1.go
}

##### AuthZEN #####
cd authzen
oapi-codegen -config config.yaml openapi.yaml
fix_code "authzen"
cd ..

##### Policies #####
cd policies
oapi-codegen -config config.yaml openapi.yaml
fix_code "policies"
cd ..

##### Attributes #####
cd attributes
oapi-codegen -config config.yaml openapi.yaml
fix_code "attributes"
cd ..

##### Authlog #####
cd authlog
oapi-codegen -config config.yaml openapi.yaml
fix_code "authlog"
cd ..

##### FSC #####
# clone repo
cd fsc
git clone -q --depth 1 https://gitlab.com/commonground/fsc/open-fsc.git
cp open-fsc/outway/authorization-interface.yaml auth/openapi.yaml

# generate
cd auth
oapi-codegen -config config.yaml openapi.yaml
cd ..

# remove repo
rm -Rf open-fsc
cd ..
