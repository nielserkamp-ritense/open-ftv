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

A Docker Compose configuration is available in the `docker` directory to run a minimal version of OpenFTV. This setup includes:
- OpenFTV PIP (Policy Information Point)
- OpenFTV PAP (Policy Administration Point)
- OpenFTV PDP (Policy Decision Point)
- OpenFTV Management Interface

To start the containers, run the following command from the project root directory:

```shell
docker compose -f docker/compose.yaml up --build
```
The following services will be available:
- PIP: http://localhost:9000
- PAP: http://localhost:9001
- PDP: http://localhost:9002
- Management Interface: http://localhost:8080
- PostgreSQL: localhost:5400 (credentials in compose file)

### FAQ

**I get the following error when starting the docker compose containers: database "openftv_pap" does not exist**

This sometimes happens when the initialization did not run. Try removing the existing postgres volume by running:
```shell
docker compose -f docker/compose.yaml down -v
```

**Warning** this will delete existing data.


## License

[Licensed under the EUPL](LICENSE.md)
