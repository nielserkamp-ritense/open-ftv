# OpenFTV

# Welcome
This Gitlab repository is part of the project Federatieve Toegangsverlening. OpenFTV is the reference implementation 
for the project.

Here you will find all code modules related to the project, test-data, utilities, scripts, etc.

The main documentation for this project (in Dutch) is found here:
https://vng-realisatie.github.io/ftv/

## Folder structure

- `oas` contains the OpenAPI specifications for the project
- `apps` contains the code for the subcomponents (e.g. `pap`, `pdp`, `pep` etc.)

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

To build everything this command from the root directory.

```shell
make all
```

### Docker Compose

A Docker Compose configuration is available in the `docker` directory to run a minimal version of OpenFTV. This setup includes:
- OpenFTV PIP (Policy Information Point)
- OpenFTV PAP (Policy Administration Point)
- OpenFTV PDP (Policy Decision Point)
- OpenFTV Management Interface

To start the containers, run the following command from the project root:

```shell
docker compose -f docker/compose.yaml up
```
The following services will be available:
- PIP: http://localhost:9000
- PAP: http://localhost:9001
- PDP: http://localhost:9002
- Management Interface: http://localhost:8080
- PostgreSQL: localhost:5400 (credentials in compose file)


## License

[Licensed under the EUPL](LICENCE.md)
