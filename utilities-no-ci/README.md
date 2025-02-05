# utilities-no-ci
This directory exists to facilitate modules with unit tests that cannot (currently) run on Gitlab Ci/CD.

## OpenSearch
Unit tests for the OpenSearch module in this directory are functional,
but only if a local deployment of OpenSearch is installed.

It is not feasible or advisable to run a full OpenSearch deployment during Gitlab CI/CD,
as it would simply take too long and not offer any added benefits.

See **opensearch_test.go** for the connection parameters used against a local OpenSearch deployment. 
