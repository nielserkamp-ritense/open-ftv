#!/bin/bash

##### AuthZEN #####
cd authzen
go tool oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Liveness & readiness #####
cd liveness
go tool oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Policies #####
cd policies
go tool oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Attributes #####
cd attributes
go tool oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Bundles #####
cd bundles
go tool oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Authlog #####
cd authlog
go tool oapi-codegen -config config.yaml openapi.yaml
cd ..

##### FSC #####
# clone repo
cd fsc
git clone -q --depth 1 https://gitlab.com/commonground/fsc/open-fsc.git
cp open-fsc/outway/authorization-interface.yaml auth/openapi.yaml

# generate
cd auth
go tool oapi-codegen -config config.yaml openapi.yaml
cd ..

# remove repo
rm -Rf open-fsc
cd ..
