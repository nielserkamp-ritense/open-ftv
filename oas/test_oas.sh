#!/bin/bash

function run {
  docker run --rm -e MAX_TEST_REQUEST_COMBINATIONS=1 -v ".:/data" --add-host host.docker.internal:host-gateway znsio/specmatic test $1 --testBaseURL=$2
}

run "/data/fsc/auth/openapi.yaml" "http://host.docker.internal:8080/v1/auth"
