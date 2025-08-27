# OpenFTV

# Welcome
This Gitlab repository is part of the project Federatieve Toegangsverlening. OpenFTV is the reference implementation 
for the project.

Here you will find all code modules related to the project, test-data, utilities, scripts, etc.

The main documentation for this project (in Dutch) is found here:
https://vng-realisatie.github.io/ftv/

## Folder structure

- `apps` contains the code for the subcomponents (e.g. `pap`, `pdp`, `pip` etc.)
- `docker` contains docker and compose related scripts
- `e2e` contains end-2-end scripts
- `eam` contains shared EAM modules
- `mock` contains mock services, including the generic mock data-service
- `oas` contains the OpenAPI specifications for the project
- `testdata` contains files with test data
- `utilities*` contains shared utility modules

## Building and running

### Cloning

```bash
git clone git@gitlab.com:digilab.overheid.nl/ecosystem/ftv/open-ftv.git
cd ftv-implementatie
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

To build and test everything, this command from the project root directory:

```shell
make all
```

### Docker Compose

A Docker Compose configuration is available in the `docker` directory to run a minimal version of OpenFTV.
This setup includes:
- several OpenFTV Managers (Authorization Management Point)
- several OpenFTV PDPs (Policy Decision Points)
- several OpenFTV Management Interfaces

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

With this setup you only get the Vlierdam services (without the Kong outway).

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
