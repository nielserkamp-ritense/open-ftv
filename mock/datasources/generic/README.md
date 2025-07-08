# Generic Mock Data-Service

# Welcome
This code module implements a generic mock data-service.

Key features:
- simple configuration of database schema using YAML or JSON (*future options: RDF turtle and/or JSON-LD*).
- simple loading of data with CSV, YAML or JSON (*future options: RDF turtle and/or JSON-LD*).
- schema and data are stored in memory; no database service(s) required.
- API endpoints to read database schema components.
- API endpoints to retrieve table data.
- custom definable endpoints; with flexible (recursive) joins.
- flexible field filtering, e.g. vertical data-minimalization.
- flexible data filtering, e.g. horizontal data-minimalization.
- flexible output options: JSON, YAML or CSV (*future options: RDF turtle and/or JSON-LD*).
- custom transformations on existing table data (or other transformations):
  - conversion: returns converted value, e.g., *string* to *int*, *int* to *string*, ...
  - comparison: returns boolean, e.g., *(not) equal*, *smaller*, *greater*, *smaller or equal*, *greater or equal*, *(not) nil*, *(not) in a list*, *(not) like*, *regex (not) matched*.
  - age: returns *age* in years from a *date of birth* field (BRP style calculation).
  - *future options*: hashing, encrypting, masking, ...

For an extensive description of these features and how to use them, see [this README](../data/README.md).

**Future options are not on the roadmap (yet).**
Please create a [Gitlab issue](https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/-/issues) if you need them.

## Building and running

```shell
git clone https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv.git
cd open-ftv
go run mock/datasources/generic/cmd/main.go --data=testdata/dataspaces/fds
```
In another shell:
```shell
curl -X POST -H'Content-Type: application/json' -H 'Accept: application/json' http://localhost:8443/v1/haalcentraal/api/brp/personen
```

## Docker images

The Gitlab CI/CD is set up to automatically create docker images in the
[Gitlab container registry](https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/container_registry/8694219).

## Configuration options

The service supports various configuration options.

Options are read and processed in the following order:
1. Configuration files.
2. Environment variables.
3. Command-line flags.

In case configuration options are defined in multiple places, only the last value will be retained and used.

The service searches for one or more configuration files with **YAML** encoding in the following locations:
- */etc/gen-ds/default.conf*
- *./etc/gen-ds.yaml* (relative to the running binary)
- *./gen-ds.yaml* (relative to the running binary)

### Configuration file (YAML)

The following is an example of a configuration file containing all possible options:
```yaml
---
svc:   # options for the API endpoint server.
  host: "<address>"             # address the service should listen on (default "0.0.0.0"; all host addresses).
  port: <port>                  # port the service should listen on (default 8443).
  tls:
    ca: "<certificate-file>"    # CA certificate to use with the service; turns on https support (no default).
    cert: "<certificate-file>"  # TLS certificate to use with the service; turns on https support (no default).
    key: "<key-file>"           # TLS private key to use with the service; turns on https support (no default).
  timeout:
    read: "<duration>"          # the read timeout for API requests (default "30s").
    write: "<duration>"         # the read timeout for API requests (default "30s").
    idle: "<duration>"          # the read timeout for API requests (default "300s").
  maxBody: <size>               # maximum allowed size of a request body (default 65536).
  cors:
    origins: "<origins>"        # CORS origins (default "*").
    headers: "<headers>"        # CORS headers (default "*").

log:   # options for the application-log.
  output: "<destination>"       # File or standard stream for writing the log (default "stdout").
  format: "<encoding>"          # Type of encoding for the log; supported are "text" and "json" (default "json").
  level: "<verbosity>"          # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
  source: true|false            # Flag to record the source location of the message in the log (default true).

data: # options for schema and table data.
  path: "<path>"                # Directory where the database-schema and table-data files can be found.
```

### Environment variables

The following is an exhaustive list of all possible environment variables the service will understand.
These match the corresponding options in a configuration file.

```text
# options for the API endpoint server.
GEN_DS_ADDRESS=<address>                       # address the service should listen on (default "0.0.0.0"; all host addresses).
GEN_DS_PORT=<port>                             # port the service should listen on (default 8443).
GEN_DS_TLA_CA=<certificate-file>               # CA certificate to use with the service; turns on https support (no default).
GEN_DS_TLS_CERT=<certificate-file>             # TLS certificate to use with the service; turns on https support (no default).
GEN_DS_TLS_KEY=<key-file>                      # TLS private key to use with the service; turns on https support (no default).
GEN_DS_READ_TIMEOUT=<duration>                 # the read timeout for API requests (default "30s").
GEN_DS_WRITE_TIMEOUT=<duration>                # the read timeout for API requests (default "30s").
GEN_DS_IDLE_TIMEOUT=<duration>                 # the read timeout for API requests (default "300s").
GEN_DS_MAX_BODY_SIZE=<size>                    # maximum allowed size of a request body (default 65536).
GEN_DS_CORS_ORIGINS=<origins>                  # CORS origins (default "*").
GEN_DS_CORS_HEADERS=<headers>                  # CORS headers (default "*").

# options for the application-log.
GEN_DS_LOG_OUTPUT=<destination>                # File or standard stream for writing the log (default "stdout").
GEN_DS_LOG_FORMAT=<encoding>                   # Type of encoding for the log; supported are "text" and "json" (default "json").
GEN_DS_LOG_LEVEL=<verbosity>                   # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
GEN_DS_LOG_SOURCE=true|false                   # Flag to record the source location of the message in the log (default true).

# options for schema and table data.
GEN_DS_DATA_PATH=<path>                        # Directory where the database-schema and table-data files can be found.
```

### Command-line flags

The following is an exhaustive list of all possible command-line flags the service will understand.
These match the corresponding options in a configuration file.

```text
# options for the API endpoint server.
--address=<address>                       # address the service should listen on (default "0.0.0.0"; all host addresses).
--port=<port>                             # port the service should listen on (default 8443).
--tls-ca=<certificate-file>               # CA certificate to use with the service; turns on https support (no default).
--tls-cert=<certificate-file>             # TLS certificate to use with the service; turns on https support (no default).
--tls-key=<key-file>                      # TLS private key to use with the service; turns on https support (no default).
--read-timeout=<duration>                 # the read timeout for API requests (default "30s").
--write-timeout=<duration>                # the read timeout for API requests (default "30s").
--idle-timeout=<duration>                 # the read timeout for API requests (default "300s").
--max-body=<size>                         # maximum allowed size of a request body (default 65536).
--cors-origins=<origins>                  # CORS origins (default "*").
--cors-headers=<headers>                  # CORS headers (default "*").

# options for the application-log.
--log-output=<destination>                # File or standard stream for writing the log (default "stdout").
--log-format=<encoding>                   # Type of encoding for the log; supported are "text" and "json" (default "json").
--log-level=<verbosity>                   # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
--log-source=true|false                   # Flag to record the source location of the message in the log (default true).

# options for schema and table data.
--data=<path>                             # Directory where the database-schema and table-data files can be found.
```

## Application log

The service writes a single application log of relevant events, so that it can aid in:
- monitoring the functioning of the service,
- debugging the root cause of problems,
- monitoring life-time events, such as displaying configuration options at startup.

Various configuration options can be used to:
- determine where the log is written,
- how to format the log,
- the verbosity level at which events will be logged.

Options for logging are processed first at startup.
This means the log is ready before any other options are checked and possible errors are printed in the log.
If logging options cause errors, these are printed using the standard Golang logging.

## License

[Licensed under the EUPL](../../../LICENSE.md)
