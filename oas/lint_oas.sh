#!/bin/bash

# Linting currently passes for ADR v1.0.0. But since ADR has been updated to v2.0.0,
# we need to update our OpenAPI specs to be compatible first.
# https://gitdocumentatie.logius.nl/publicatie/api/adr/<version>/media/linter.yaml)

spectral lint -r .spectral.yml liveness/openapi.yaml
spectral lint -r .spectral.yml attributes/openapi.yaml
spectral lint -r .spectral.yml policies/openapi.yaml
spectral lint -r .spectral.yml bundles/openapi.yaml
spectral lint -r .spectral.yml authlog/openapi.yaml
# spectral lint -r .spectral.yml authzen/openapi.yaml
spectral lint -r .spectral.yml schema/openapi.yaml
