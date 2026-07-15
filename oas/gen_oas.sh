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
cd fsc/auth
go tool oapi-codegen -config config.yaml openapi.yaml

# remove repo
rm -Rf open-fsc
cd ..
