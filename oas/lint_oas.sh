#!/bin/bash

docker run --rm -it -v .:/tmp stoplight/spectral lint -r "/tmp/.spectral.yml" "/tmp/attributes/oas-attributes.yaml"
docker run --rm -it -v .:/tmp stoplight/spectral lint -r "/tmp/.spectral.yml" "/tmp/policies/oas-policies.yaml"
docker run --rm -it -v .:/tmp stoplight/spectral lint -r "/tmp/.spectral.yml" "/tmp/authzen/oas-authzen.yaml"
docker run --rm -it -v .:/tmp stoplight/spectral lint -r "/tmp/.spectral.yml" "/tmp/fds/ledenlijst/oas-leden.yaml"
docker run --rm -it -v .:/tmp stoplight/spectral lint -r "/tmp/.spectral.yml" "/tmp/fsc/auth/oas-auth.yaml"
