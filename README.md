# OpenFTV

This repository contains the reference implementation of [Federatieve Toegangsverlening](https://vng-realisatie.github.io/ftv/).

## Folder structure

- `apps      ` code for EAM services (e.g. `pap`, `pdp`, `pip` etc.).
- `demos     ` demo & presentation materials.
- `docker    ` docker and compose related scripts.
- `e2e       ` end-2-end scripts.
- `eam       ` shared generic EAM components.
- `external  ` OpenFTV plugins for external software products. 
- `migrations` database migration scripts.
- `mock      ` mock services, including the generic mock data-service.
- `oas       ` OpenAPI specifications for the project.
- `resources ` various resources; e.g., project icon.
- `scripts   ` various internal scripts.
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

### Docker Compose

Docker Compose configurations are available in the `docker` directory to run local versions of OpenFTV.
These setups include:
- OpenFTV Manager (Authorization Management: PAP, PIP, tag management, bundle management, database migrations)
- OpenFTV PDP
- OpenFTV Management Interface
- Kong gateway with OpenFTV AuthZEN plugin
- generic mock data-service

#### Extended test setup

To start a full test-setup with multiple organizations and connecting gateways,
run the following command from the project root directory:

```shell
docker compose -f docker/compose.yaml up --build
```
The following services will be available:
- Gemeente Vlierdam:
  - Management Interface: http://localhost:8080
  - Authorization Manager (PAP+PIP): http://localhost:9000
  - Outway (Kong): http://localhost:9002
  - PDP1: http://localhost:9004
  - mock dataspace: http://localhost:9010
  - Postgres: http://localhost:5400 (credentials in compose file)
- RvIG:
  - Management Interface: http://localhost:8082
  - Authorization Manager (PAP+PIP): http://localhost:9020
  - Inway (Kong): http://localhost:9022
  - PDP1: http://localhost:9024
  - mock dataspace: http://localhost:9030
  - Postgres: http://localhost:5420 (credentials in compose file)
- RDW:
  - Management Interface: http://localhost:8084
  - Authorization Manager (PAP+PIP): http://localhost:9040
  - Inway (Kong): http://localhost:9042
  - PDP1: http://localhost:9044
  - mock dataspace: http://localhost:9050
  - Postgres: http://localhost:5440 (credentials in compose file)

#### Reduced test setup

Alternatively, you can use the following command to run a simple local setup with just gemeente Vlierdam:

```shell
docker compose -f docker/compose-vlierdam.yaml up --build
```

With this setup you get a simple test-setup with just the Vlierdam services, 
without the Kong gateway as it is only needed for external connections.

#### OpenFGA demo

Run:

```shell
docker compose -f docker/openfga-playground.yaml up --build
```

And point your browser to http://localhost:3000/playground.

#### FAQ

**I get the following error when starting the docker compose containers: database "openftv" does not exist**

This sometimes happens when the initialization script did not run properly.
Try removing the existing postgres volume by running:
```shell
docker compose -f docker/compose.yaml down -v
```

**Warning** this will delete existing data.
