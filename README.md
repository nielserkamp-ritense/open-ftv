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

To build everything this command from the root directory

```shell
make all
```

## License

[Licensed under the EUPL](LICENCE.md)
