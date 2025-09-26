#!/bin/bash

##### AuthZEN #####
cd authzen
oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Liveness & readiness #####
cd liveness
oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Policies #####
cd policies
oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Attributes #####
cd attributes
oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Bundles #####
cd bundles
oapi-codegen -config config.yaml openapi.yaml
cd ..

##### Authlog #####
cd authlog
oapi-codegen -config config.yaml openapi.yaml
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
