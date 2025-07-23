#!/bin/bash

curl -s https://apis.developer.overheid.nl/static/adr/ruleset.yaml > .spectral.yml

spectral lint -r .spectral.yml attributes/openapi.yaml
spectral lint -r .spectral.yml policies/openapi.yaml
spectral lint -r .spectral.yml bundles/openapi.yaml
spectral lint -r .spectral.yml authzen/openapi.yaml
