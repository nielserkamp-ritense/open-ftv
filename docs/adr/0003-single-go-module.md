# Single Go module for the whole repository

The repository is **one Go module** (`gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv`,
root `go.mod`), not a collection of per-component modules. Before this decision the repo
held 35 modules (every `eam/*`, `apps/*`, `utilities*`, `oas`, `migrations`, `mock/*` had
its own `go.mod`), stitched together by **three overlapping resolution mechanisms**: a root
`go.work` for local builds, relative `replace` blocks in every `go.mod` for Docker builds,
and `require` pseudo-versions pinning published commits for a publish-for-external-reuse
story (plus `remove_replace.sh` helpers to strip mechanism two so mechanism three could
take over).

We merged to one module because the split delivered none of its intended benefits:

- The pinned pseudo-versions were **incoherent** (within one app they pointed at internal
  commits five months apart) and **never exercised** — `replace` and `go.work` always won,
  so the publish path was untested and almost certainly broken.
- Nothing outside this repository consumes the modules; `remove_replace.sh` was referenced
  by nothing.
- The repo already versions as **one unit**: global `vX.Y.Z` git tags drive releases and
  image tags. There were never per-component tags.
- Every structural change (new component, new dependency) required updating four to five
  hand-maintained lists (`go.work`, `require`, `replace`, makefile `DIRS`, CI test list),
  which had already drifted apart.

## Consequences

- **Import paths are unchanged.** Every former module path was exactly
  `<root>/<directory>`, so the single root module keeps all imports valid.
- **Deliberately no `/v2` suffix on the module path**, although `v2.x` tags exist. Go's
  semantic-import-versioning rule means external `go get` of this module at `v2+` tags
  does not work. That is accepted: the tags version releases and container images, not a
  published Go API, and nothing external imports the module. Do not "fix" this by adding
  the suffix — it would force rewriting every import in the repo for no consumer.
- **One dependency graph.** External dependency versions are unified via MVS in the root
  `go.mod`; components can no longer pin diverging versions of the same dependency.
  (In practice `go.work` already unified them for local and CI builds; only Docker builds
  ever saw per-component resolution.)
- **Tests that need external services are excluded by build tag**, not by module boundary:
  files under `utilities-no-ci/` that require OpenSearch or GitHub credentials carry
  `//go:build external` and run via `make test-external`. `go test ./...` is the whole
  suite everywhere (local, CI, any directory).
- **Docker builds copy the repo root** (`COPY go.mod go.sum` + `go mod download`, then
  `COPY . .`) instead of hand-picked component directories, so adding an internal
  dependency can no longer break only the Docker build.
- Re-splitting into per-component modules is expensive; if external reuse of a component
  ever becomes real, extracting that component into its own repository is the more likely
  path than re-splitting this one.
