# OpenFTV

# Welcome
This Gitlab repository is part of the project Federatieve Toegangsverlening. OpenFTV is the reference implementation 
for the project.

Here you will find all code modules related to the project, test-data, utilities, scripts, etc.

The repo is set up as a mono-repo. This allows separate modules to be re-used in other projects.

The main documentation for the FTV project (in Dutch) can be found at
https://vng-realisatie.github.io/ftv/.

## Folder structure

- `apps` code for EAM services (e.g. `pap`, `pdp`, `pip` etc.)
- `docker` docker and compose related scripts
- `e2e` end-2-end scripts
- `eam` shared generic EAM modules
- `migrations` database migration scripts
- `mock` mock services, including the generic mock data-service
- `oas` OpenAPI specifications for the project
- `testdata` files with test data
- `utilities*` shared generic utility modules

## Building and running

### Cloning

```bash
git clone git@gitlab.com:digilab.overheid.nl/ecosystem/ftv/open-ftv.git
cd open-ftv
```

### Development setup

Install the following tools

- [Docker Desktop / Docker engine](https://docs.docker.com/install/)
- [Golang](https://golang.org/doc/install)
- [Spectral](https://stoplight.io/open-source/spectral)
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
- htpasswd
- Make

### Building

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

Alternatively, you can use the following command to run a simple local setup for just gemeente Vlierdam:

```shell
docker compose -f docker/compose-vlierdam.yaml up --build
```

With this setup you get a simple test-setup with just the Vlierdam services, 
without the Kong outway (as it's not needed).

### FAQ

**I get the following error when starting the docker compose containers: database "openftv" does not exist**

This sometimes happens when the initialization script did not run properly.
Try removing the existing postgres volume by running:
```shell
docker compose -f docker/compose.yaml down -v
```

**Warning** this will delete existing data.

## License

[Licensed under the EUPL](LICENSE.md)
