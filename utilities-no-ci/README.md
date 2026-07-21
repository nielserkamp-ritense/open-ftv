# utilities-no-ci

This directory holds utility packages whose unit tests need external services and
therefore cannot run unattended in GitLab CI/CD. Those test files carry the
`//go:build external` build tag, so a plain `go test ./...` (locally and in CI) skips
them. Run them deliberately with:

```shell
make test-external          # = go test -tags external ./utilities-no-ci/...
```

Note: `module/version_test.go` needs no external service, is untagged, and runs in CI
like any other test. A rename of this directory to `utilities-external` (matching the
build tag, now that CI exclusion no longer works via module boundaries) is planned as a
follow-up MR.

## OpenSearch

Unit tests for the OpenSearch package are functional, but only against a local
OpenSearch deployment. Running a full OpenSearch deployment during GitLab CI/CD would
simply take too long and not offer any added benefits.

See **opensearch_test.go** for the connection parameters used against a local OpenSearch
deployment.
