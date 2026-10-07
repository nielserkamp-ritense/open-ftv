# OpenFTV

This repository contains the reference implementation of [Federatieve Toegangsverlening](https://vng-realisatie.github.io/ftv/).

## Folder structure

- `apps      ` code for EAM services (e.g. `pap`, `pdp`, `pip`, `pep-kong` etc.).
- `demo      ` configuration of the Laadpalen demo, per organization.
- `docker    ` docker and compose related scripts.
- `e2e       ` end-to-end tests of the Laadpalen demo.
- `eam       ` shared generic EAM components.
- `migrations` database migration scripts.
- `mock      ` mock services, including the generic mock data-service.
- `oas       ` OpenAPI specifications for the project.
- `resources ` various resources; e.g., project icon.
- `testdata  ` files with test data.
- `utilities*` shared generic utility components.

## Development setup

How we write commits, code and merge requests is in [CONTRIBUTING.md](CONTRIBUTING.md).

### Prerequisites

- [mise](https://github.com/jdx/mise)
- [Docker](https://docs.docker.com/install/)
- htpasswd

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
root:

```shell
go generate ./oas/...   # regenerate the Go code of the OpenAPI specs
go vet ./...
golangci-lint run       # as the Lint Golang job in CI does
go test ./...
```

### Local setup

The local setup runs the Proefopstelling of the Laadpalen application with three organizations: Gemeente Vlierdam,
RvIG and RDW. It is split in two parts:

- `docker/compose.yaml` runs the infrastructure: PostgreSQL, and a Kong gateway and Keycloak per organization.
- `modd.conf` runs the OpenFTV components and the mock data-services on the host, and restarts them when their code
  changes. The Go services run under [Delve](https://github.com/go-delve/delve), so a debugger can attach to them.

The demo's configuration per organization (policies, tags, bundles, PDP pull-configs and Kong) is in
`demo/<organization>`. The Kong configs, bundles and PDP pull-configs are templates (`*.template`): their addresses
are variables, set in `demo/addresses.env`. The `render-demo` service in `docker/compose.yaml` fills each one in into
the file next to it without `.template`, where the gateways and modd read it; modd runs it first whenever it starts.
After changing a template or an address, run
`docker compose -f docker/compose.yaml up -d --force-recreate render-demo vlierdam-api-gateway rvig-api-gateway rdw-api-gateway`
and restart modd.

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

On every start, modd saves each Policy Editor's header title and colors through the settings API (see the end of
`modd.conf`), replacing changes made on the Instellingen page.

modd does not reload `modd.conf` itself: restart modd after changing it.

#### End-to-end tests

The end-to-end tests of the Laadpalen demo run against this setup. See [e2e/README.md](e2e/README.md).

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

See [CONTRIBUTING.md](CONTRIBUTING.md#commits).

##### Creating a release

OpenFTV is pre-1.0, so a breaking change bumps the *minor* and everything else the *patch* (see `cliff.toml`).

A release is made by starting the manual `release` job on a `main` pipeline. It derives the version and the release
notes from the commits since the last tag, creates the annotated tag and the GitLab Release, and the tag pipeline
publishes it. To release a different version than the computed one, set `RELEASE_TAG` (for example `v0.3.0`) in the
variables box when starting the job. It must look like `v1.2.3`, must not exist yet and must be newer than the latest
tag. The job only becomes available once the rest of the pipeline has passed.

The job authenticates with the `RELEASE_TOKEN` CI variable, because a tag created with the CI job token does not start
a pipeline. It must be a group access token on the `ecosystem/ftv` group (Maintainer, `api` scope): the tag pipeline
pulls its images through the group's dependency proxy, which answers "not found" to a project access token's bot.

The pre-release version on `main` always assumes a *patch* bump, so it is numbered
`0.2.1-main.<n>` even when the next release will be `0.3.0`. That is expected.

#### Simulatieomgeving

There are two independent deployments running on the Simulatieomgeving, each deployed automatically by Flux. Both run
the Laadpalen demo.

The latest `main` **pre-release**:

- Laadpalen application: https://openftv-vlierdam-pre.simulatie.datastelsel.nl
- Management interfaces:
  - Gemeente Vlierdam: https://openftv-vlierdam-beheer-pre.simulatie.datastelsel.nl
  - RvIG (BRP): https://openftv-brp-beheer-pre.simulatie.datastelsel.nl
  - RDW: https://openftv-rdw-beheer-pre.simulatie.datastelsel.nl
- Keycloak: `https://openftv-<vlierdam|brp|rdw>-keycloak-pre.simulatie.datastelsel.nl`
- Namespace: `fds-tenant-ictu-open-ftv`
- Manifests: `deploy/kustomize/simulation-pre-release`

The latest tagged **release**:

- Laadpalen application: https://openftv-vlierdam.simulatie.datastelsel.nl
- Management interfaces:
  - Gemeente Vlierdam: https://openftv-vlierdam-beheer.simulatie.datastelsel.nl
  - RvIG (BRP): https://openftv-brp-beheer.simulatie.datastelsel.nl
  - RDW: https://openftv-rdw-beheer.simulatie.datastelsel.nl
- Keycloak: `https://openftv-<vlierdam|brp|rdw>-keycloak.simulatie.datastelsel.nl`
- Namespace: `fds-organization-gemeente-vlierdam`
- Manifests: `deploy/kustomize/simulation-release`

[Tenant configuration](https://gitlab.com/datastelsel.nl/federatief/simulation/platform/-/tree/main/flux/tenants/simulation/ictu-open-ftv)

Every deploy resets the demo: the `demo-reset-<version>` Job empties each manager's database, so edits made in the
management interfaces are lost, and the managers seed the demo policies from `demo/` again. A new file in
`demo/<organization>/manager/policies` also needs a line in both overlays' `kustomization.yaml`. The Laadpalen
application reaches Vlierdam's mock dataspace, outway and PDP1 through `/dataspace`, `/outway` and `/pdp1` on its own
host.

The overlays read the demo's bundle configurations, PDP pull-configs and Kong configurations from `demo/`; the
deploy job fills in their addresses from the overlay's `addresses.env`. A new variable in a template needs a value
in both overlays' `addresses.env` as well as in `demo/addresses.env`.

[Cluster access](https://gitlab.com/datastelsel.nl/federatief/simulation/simulation/#access-to-k8s-cluster)

Check the deployed images:

```shell
kubectl get pods -n fds-tenant-ictu-open-ftv -o jsonpath='{..image}'
kubectl get pods -n fds-organization-gemeente-vlierdam -o jsonpath='{..image}'
```

##### Header title and colors

Each management interface takes its header title and colors from the settings API. They are in
`demo/<organization>/manager/settings.json`. The reset Job sets them on every deploy.

##### Accounts

| Username       | Password   | Role      |
| -------------- | ---------- | --------- |
| `admin-user`   | `password` | `admin`   |
| `author-user`  | `password` | `author`  |
| `auditor-user` | `password` | `auditor` |
