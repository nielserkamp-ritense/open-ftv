# OpenFTV

This repository contains the reference implementation of [Federatieve Toegangsverlening](https://vng-realisatie.github.io/ftv/).

## Folder structure

- `apps      ` code for EAM services (e.g. `pap`, `pdp`, `pip` etc.).
- `demos     ` demo & presentation materials.
- `docker    ` docker and compose related scripts.
- `eam       ` shared generic EAM components.
- `external  ` OpenFTV plugins for external software products. 
- `migrations` database migration scripts.
- `mock      ` mock services, including the generic mock data-service.
- `oas       ` OpenAPI specifications for the project.
- `resources ` various resources; e.g., project icon.
- `testdata  ` files with test data.
- `utilities*` shared generic utility components.

## Development setup

### Prerequisites

- [mise](https://github.com/jdx/mise)
- [Docker](https://docs.docker.com/install/)
- htpasswd
- Make

### Installation

```bash
git clone git@gitlab.com:digilab.overheid.nl/ecosystem/ftv/open-ftv.git
cd open-ftv
```

The tooling is defined in `mise.toml`.

```bash
mise trust
mise install
```

### Building

The repository is a single Go module (see
[ADR 0003](docs/adr/0003-single-go-module.md)), so the standard Go commands work from the
root: `go build ./...`, `go test ./...`.

To build and test everything locally, run the following command from the project root directory:

```shell
make all
```

### Local setup

The local setup runs the Proefopstelling of the Laadpalen application with three organizations: Gemeente Vlierdam,
RvIG and RDW. It is split in two parts:

- `docker/compose.yaml` runs the infrastructure: PostgreSQL, and a Kong gateway and Keycloak per organization.
- `modd.conf` runs the OpenFTV components and the mock data-services on the host, and restarts them when their code
  changes. The Go services run under [Delve](https://github.com/go-delve/delve), so a debugger can attach to them.

Besides the prerequisites above, this needs [modd](https://github.com/cortesi/modd), Delve and Node.js. Run
`npm install` once in `apps/management-interface` and `apps/laadpalen-applicatie`.

Start both from the project root directory, each in its own terminal:

```shell
docker compose -f docker/compose.yaml up --build
modd
```

The following services will be available:
- Gemeente Vlierdam:
  - Laadpalen application: http://localhost:8080
  - Policy Editor: http://localhost:5173
  - Authorization Manager (PAP+PIP): http://localhost:9000
  - Outway (Kong): http://localhost:9002
  - PDP1: http://localhost:9004
  - PDP2: http://localhost:9005
  - mock dataspace: http://localhost:9010
  - Keycloak: http://localhost:8088
- RvIG:
  - Policy Editor: http://localhost:5174
  - Authorization Manager (PAP+PIP): http://localhost:9020
  - Inway (Kong): http://localhost:9022
  - PDP1: http://localhost:9024
  - mock dataspace: http://localhost:9030
  - Keycloak: http://localhost:8090
- RDW:
  - Policy Editor: http://localhost:5175
  - Authorization Manager (PAP+PIP): http://localhost:9040
  - Inway (Kong): http://localhost:9042
  - PDP1: http://localhost:9044
  - mock dataspace: http://localhost:9050
  - Keycloak: http://localhost:8100
- PostgreSQL: localhost:5432 (credentials in `docker/compose.yaml`), with a database per component.

modd does not reload `modd.conf` itself: restart modd after changing it.

#### FAQ

**A component fails to start because its database does not exist**

The databases are created by `docker/scripts/postgres-setup.sh`, which only runs when the postgres volume is empty.
Try removing the existing postgres volume by running:
```shell
docker compose -f docker/compose.yaml down -v
```

**Warning** this will delete existing data.

#### Release

Commits on a feature-branch will only run the test-jobs.

Everything is published under `registry.gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/<component>` for container
images and `.../open-ftv/charts/<chart>` for Helm charts.

Commits on the `main` branch publish a SemVer *pre-release*, `<next-version>-<branch>.<commits since the last tag>`. 
For example chart `0.1.1-main.16` and image `v0.1.1-main.16`.

Tags publish the tag itself, for example chart `0.1.1` and image `v0.1.1`, and additionally move `latest`.

Helm chart versions are bare SemVer, while container images carry a `v` prefix.

##### Commit messages

Commits follow [Conventional Commits](https://www.conventionalcommits.org), e.g. `feat: add a Policy editor` or
`fix: answer a rejected bearer token with 401`. The version bump and the release notes are derived from them, and
the `Lint commit messages` job checks every commit on a branch. Check your branch locally with
`npx --package @commitlint/cli --package @commitlint/config-conventional commitlint --from origin/main`.

- `feat` and `fix` end up in the release notes, as do `perf`, `refactor`, `docs`, `build`, `ci` and `revert`. The
  `Revert "..."` message of `git revert` and GitLab's Revert button counts as a `revert`.
- `chore`, `style` and `test` are left out, and a range with only those has nothing to release. A breaking one is
  listed under *Other*.
- Mark a breaking change with `!` (`feat!: ...`) or with a `BREAKING CHANGE:` footer. Both bump the *minor* while
  pre-1.0. A footer needs a blank line above it, which `git commit -m "feat: ..." -m "BREAKING CHANGE: ..."` gives
  you. The linter rejects a footer without that blank line, because the release notes would otherwise drop the
  commit or miss that it is breaking.
- Keep the subject lowercase. `feat: Add a thing` is rejected.

##### Creating a release

OpenFTV is pre-1.0, so a breaking change bumps the *minor* and everything else the *patch* (see `cliff.toml`).

A release is made by starting the manual `release` job on a `main` pipeline. It derives the version and the release
notes from the commits since the last tag, creates the annotated tag and the GitLab Release, and the tag pipeline
publishes it. To release a different version than the computed one, set `RELEASE_TAG` (for example `v0.3.0`) in the
variables box when starting the job. It must look like `v1.2.3`, must not exist yet and must be newer than the latest
tag. The job only becomes available once the rest of the pipeline has passed.

The job authenticates with the `RELEASE_TOKEN` CI variable, a project access token (Maintainer, `api` scope), because
a tag created with the CI job token does not start a pipeline.

The pre-release version on `main` always assumes a *patch* bump, so it is numbered
`0.2.1-main.<n>` even when the next release will be `0.3.0`. That is expected.

#### Simulatieomgeving

There are two independent deployments running on the Simulatieomgeving, each deployed automatically by Flux.

The latest `main` **pre-release**:

- Management interface: https://open-ftv.simulatie.datastelsel.nl
- Namespace: `fds-tenant-ictu-open-ftv`
- Manifests: `deploy/kustomize/simulation-pre-release`
- [Tenant configuration](https://gitlab.com/datastelsel.nl/federatief/simulation/platform/-/tree/main/flux/tenants/simulation/ictu-open-ftv)

The latest tagged **release**:

- Management interface: https://open-ftv-vlierdam.simulatie.datastelsel.nl
- Namespace: `fds-organization-gemeente-vlierdam`
- Manifests: `deploy/kustomize/simulation-release`
- [Tenant configuration](https://gitlab.com/datastelsel.nl/federatief/simulation/platform/-/tree/main/flux/tenants/simulation/ictu-open-ftv)

[Cluster access](https://gitlab.com/datastelsel.nl/federatief/simulation/simulation/#access-to-k8s-cluster)

Check the deployed images:

```shell
kubectl get pods -n fds-tenant-ictu-open-ftv -o jsonpath='{..image}'
kubectl get pods -n fds-organization-gemeente-vlierdam -o jsonpath='{..image}'
```

##### Accounts

| Username       | Password   | Role      |
| -------------- | ---------- | --------- |
| `admin-user`   | `password` | `admin`   |
| `author-user`  | `password` | `author`  |
| `auditor-user` | `password` | `auditor` |
