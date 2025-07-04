# OpenFTV - FSC Authorization Plugin

# Welcome
This code module implements an authorization plugin for the OpenFSC Inway and Outway services.

It supports the following interfaces:
- API endpoint to handle authorization requests (AuthZEN Evaluation API).
- API endpoint to handle authorization requests (deprecated OpenFSC Authorization API).
- API endpoints to push attributes; intended for external PIP systems, such as HR, IAM, etc.
- functionality to pull attributes from external PIPs.

## Building and running

See the [README](../../README.md) in the top-level directory.

## Docker images

The Gitlab CI/CD is set up to automatically create docker images in the
[Gitlab container registry](https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/container_registry/8465187).

## Configuration options

The service supports various configuration options.

Options are read and processed in the following order:
1. Configuration files.
2. Environment variables.
3. Command-line flags.

In case configuration options are defined in multiple places, only the last value will be retained and used.

The service searches for one or more configuration files with **YAML** encoding in the following locations:
- */etc/pip/default.conf*
- *./etc/pip.yaml* (relative to the running binary)
- *./pip.yaml* (relative to the running binary)

### Configuration file (YAML)

The following is an example of a configuration file containing all possible options:
```yaml
---
svc:   # options for the API endpoint server.
  host: "<address>"             # address the service should listen on (default "0.0.0.0"; all host addresses).
  port: <port>                  # port the service should listen on (default 8080).
  tls:
    ca: "<certificate-file>"    # CA certificate to use with the service; turns on https support (no default).
    cert: "<certificate-file>"  # TLS certificate to use with the service; turns on https support (no default).
    key: "<key-file>"           # TLS private key to use with the service; turns on https support (no default).
  timeout:
    read: "<duration>"          # the read timeout for API requests (default "30s").
    write: "<duration>"         # the read timeout for API requests (default "30s").
    idle: "<duration>"          # the read timeout for API requests (default "300s").
  maxBody: <size>               # maximum allowed size of a request body (default 65536).

log:   # options for the application-log.
  output: "<destination>"       # File or standard stream for writing the log (default "stdout").
  format: "<encoding>"          # Type of encoding for the log; supported are "text" and "json" (default "json").
  level: "<verbosity>"          # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
  source: true|false            # Flag to record the source location of the message in the log (default true).

policies:   # policies used for authorizing FSC requests.
  language: "<language>"        # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
  store:
    path: "<directory>"         # Location on disk where static policy files can be found (no default).
    recurse: true|false         # Flag to indicate the given directory and its subdirectories must be search recursively for policy files (default false).

pip:   # attributes used for authorizing FSC requests.
  store:
    path: "<directory>"         # Location on disk where static attribute files can be found (no default).
    recurse: true|false         # Flag to indicate the given directory and its subdirectories must be search recursively for attribute files (default false).
  pull:   # see the section "Pull configurations" below for more information.
    configPath: "<file>"        # Location on disk where a "pull configuration" file can be found (no default).

cerbos:   # options for connecting with a cerbos PDP (sidecar) for authorizing FSC requests.
  address: "<address>"          # Address of the Cerbos API.
  ca: "<file>"                  # CA certificate to use with the Cerbos APIs.
  admin:
    address: "<address>"        # Address of the Cerbos admin API.
    user: "<user>"              # User-id to authenticate with the Cerbos admin API.
    password: "<user>"          # User-id to authenticate with the Cerbos admin API.
```

### Environment variables

The following is an exhaustive list of all possible environment variables the service will understand.
These match the corresponding options in a configuration file.

```text
# options for the API endpoint server.
FSC_AUTH_ADDRESS=<address>                       # address the service should listen on (default "0.0.0.0"; all host addresses).
FSC_AUTH_PORT=<port>                             # port the service should listen on (default 8080).
FSC_AUTH_TLA_CA=<certificate-file>               # CA certificate to use with the service; turns on https support (no default).
FSC_AUTH_TLS_CERT=<certificate-file>             # TLS certificate to use with the service; turns on https support (no default).
FSC_AUTH_TLS_KEY=<key-file>                      # TLS private key to use with the service; turns on https support (no default).
FSC_AUTH_READ_TIMEOUT=<duration>                 # the read timeout for API requests (default "30s").
FSC_AUTH_WRITE_TIMEOUT=<duration>                # the read timeout for API requests (default "30s").
FSC_AUTH_IDLE_TIMEOUT=<duration>                 # the read timeout for API requests (default "300s").
FSC_AUTH_MAX_BODY_SIZE=<size>                    # maximum allowed size of a request body (default 65536).

# options for the application-log.
FSC_AUTH_LOG_OUTPUT=<destination>                # File or standard stream for writing the log (default "stdout").
FSC_AUTH_LOG_FORMAT=<encoding>                   # Type of encoding for the log; supported are "text" and "json" (default "json").
FSC_AUTH_LOG_LEVEL=<verbosity>                   # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
FSC_AUTH_LOG_SOURCE=true|false                   # Flag to record the source location of the message in the log (default true).

# policies used for authorizing FSC requests.
FSC_AUTH_POLICIES_LANGUAGE=<language>            # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
FSC_AUTH_POLICIES_STORE=<directory>              # Location on disk where static policy files can be found (no default).
FSC_AUTH_POLICIES_STORE_RECURSE=true|false       # Flag to indicate the given directory and its subdirectories must be search recursively for policy files (default false).

# attributes used for authorizing FSC requests.
FSC_AUTH_FSC_AUTH_STORE=<directory>                   # Location on disk where static attribute files can be found (no default).
FSC_AUTH_FSC_AUTH_STORE_RECURSE=true|false            # Flag to indicate the given directory and its subdirectories must be search recursively for attribute files (default false).
FSC_AUTH_PULL_CONFIGS=<file>                     # Location on disk where a "pull configuration" file can be found (no default).

# options for connecting with a cerbos PDP (sidecar) for authorizing FSC requests.
FSC_AUTH_CERBOS_ADDRESS=<address>                # Address of the Cerbos API.
FSC_AUTH_CERBOS_ADMIN=<address>                  # Address of the Cerbos admin API.
FSC_AUTH_CERBOS_USER=<user>                      # User-id to authenticate with the Cerbos admin API.
FSC_AUTH_CERBOS_PSWD=<user>                      # User-id to authenticate with the Cerbos admin API.
FSC_AUTH_CERBOS_CA=<file>                        # CA certificate to use with the Cerbos APIs.
```

### Command-line flags

The following is an exhaustive list of all possible command-line flags the service will understand.
These match the corresponding options in a configuration file.

```text
# options for the API endpoint server.
--address=<address>                       # address the service should listen on (default "0.0.0.0"; all host addresses).
--port=<port>                             # port the service should listen on (default 8080).
--tls-ca=<certificate-file>               # CA certificate to use with the service; turns on https support (no default).
--tls-cert=<certificate-file>             # TLS certificate to use with the service; turns on https support (no default).
--tls-key=<key-file>                      # TLS private key to use with the service; turns on https support (no default).
--read-timeout=<duration>                 # the read timeout for API requests (default "30s").
--write-timeout=<duration>                # the read timeout for API requests (default "30s").
--idle-timeout=<duration>                 # the read timeout for API requests (default "300s").
--max-body=<size>                         # maximum allowed size of a request body (default 65536).

# options for the application-log.
--log-output=<destination>                # File or standard stream for writing the log (default "stdout").
--log-format=<encoding>                   # Type of encoding for the log; supported are "text" and "json" (default "json").
--log-level=<verbosity>                   # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
--log-source=true|false                   # Flag to record the source location of the message in the log (default true).

# policies used for authorizing FSC requests.
--policies-language=<language>            # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
--policies-store=<directory>              # Location on disk where static policy files can be found (no default).
--policies-store-recurse=true|false       # Flag to indicate the given directory and its subdirectories must be search recursively for policy files (default false).

# attributes used for authorizing FSC requests.
--pip-store=<directory>                   # Location on disk where static attribute files can be found (no default).
--pip-store-recurse=true|false            # Flag to indicate the given directory and its subdirectories must be search recursively for attribute files (default false).
--pip-pull-configs=<file>                 # Location on disk where a "pull configuration" file can be found (no default).

# options for connecting with a cerbos PDP (sidecar) for authorizing FSC requests.
--cerbos-address=<address>                # Address of the Cerbos API.
--cerbos-admin=<address>                  # Address of the Cerbos admin API.
--cerbos-user=<user>                      # User-id to authenticate with the Cerbos admin API.
--cerbos-pswd=<user>                      # User-id to authenticate with the Cerbos admin API.
--cerbos-ca=<file>                        # CA certificate to use with the Cerbos APIs.
```

### Pull configurations

A pull configuration is used to schedule pulling attributes from external PIP systems, such as IAM, HR, etc.
Retrieved attributes are cached in the PIP of the service.
The cache is refreshed after a certain amount of time according to the given schedule.

The location of this file is governed by the pip/pull/configPath option in the general configuration.
Its use is optional.

A pull configuration is a **YAML** encoded file with the following layout:
```yaml
description: "<text>"                  # optional description.
sources:                               # list of sources to pull data from.
  - name: "<text>"                     # unique key for the source.
    description: "<text>"              # optional description.
    requests:                          # list of data pull request configurations.
      - name: "<text>"                 # unique key for the request.
        description: "<test>"          # optional description.
        method: "<method>"             # http method to use for the request.
        uri: "<uri>"                   # uri of the API request handler.
        headers:                       # optional list of headers to add to the request.
          - "<key>": "<value>"
        contentType: "<text>"          # optional content-type for a request body; supports YAML and JSON (default is JSON).
        timeout: "<duration>"          # timeout for the request.
        tlsCA: "<file>"                # CA certificate file to use for https requests.
        tlsCert: "<file>"              # TLS certificate file to use for https requests.
        tlsKey: "<file>"               # TLS private key file to use for https requests.
        insecure: true|false           # flag to skip server certificate validation (*1).
        parameters:                    # optional list of parameters for the request.
          - name: "<text>"             # unique key of the parameter.
            in: "<location>"           # location for the parameter; supported are "path", "query" and "body" (*2).
            description: "<text>"      # optional description.
            type: "<type>"             # type of parameter; supports "string", "integer", "float", "boolean".
            value: <any>               # value of the parameter; mutually exclusive with attribute.
            attribute: "<key>"         # key of an existing attribute to use as the value; mutually exclusive with value.
        interval: "<duration>"         # interval time between subsequent requests (*3).
        initialInterval: "<duration>"  # time before the first request (*3).
        schedule: "<crontab-rule>"     # a crontab inspired schedule (*3).
        mapping:                       # mapping to decode the result into attributes (*4).
```

(*1) This option makes the https connection insecure by default. DO NOT USE IN PRODUCTION!

(*2) Parameters for the "path" replace placeholders in the URI marked with two colon characters; e.g., *:param1:*, *:param2:*.
For this to work, the **name** of the parameter must match the identifier between the two colons.
Parameters for the "query" are added as query parameters to the URI.
Parameters for the "body" are structured into an object and encoded as JSON or YAML, according to the **contentType** option.

(*3) Options to describe the schedule or interval.
Either **interval** (with an optional **initialInterval**) or **schedule** must be defined.
If **interval** is defined, the request will be executed repeatedly, with the given **interval** between calls.
If the **schedule** rule is defined, the request will be executed according to standard crontab rules.

(*4) see [response.go](../../eam/pip/network/response.go) for details of the mapping mechanism.

Note that pull configurations may also be encoded in JSON or TOML format.
However, for brevity, we will only include the YAML layout above, as a JSON or TOML layout can be inferred from it.

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

[Licensed under the EUPL](../../LICENSE.md)
