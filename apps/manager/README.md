# OpenFTV - Manager (PAP + PIP)

# Welcome
This code module implements a Policy Information Point (PIP) and Policy Administration Point (PAP) in a single service.

It supports the following interfaces:
- API endpoints to manage policies; intended for userinterfaces.
- API endpoints to manage attributes, entities and relations; intended for userinterfaces.
- API endpoints to push attributes; intended for external PIP systems, such as HR, IAM, etc.
- functionality to pull attributes from external PIPs.
- *TODO*: API endpoint to retrieve a batch of policies; intended for PDPs.
- *TODO*: functionality to push a batch of policies to a PDP.
- *TODO*: functionality to push a policy or batch of policies to a git repository.
- *TODO*: functionality to push a batch of attributes, entities and/or relations to a PDP.

## Building and running

See the [README](../../README.md) in the top-level directory.

## Docker images

The Gitlab CI/CD is set up to automatically create docker images in the
[Gitlab container registry](https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/container_registry/8582423).

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

persist:   # this is where the Manager stores policies, attributes, entities and relations; see the section "About persistence" below for examples.
  type: "<type>"                # Type of persistence backend; supported are "etcd", "consul", "postgres" (no default).
  addresses: "<addresses>"      # One or more addresses of persistence services (no default).
  base: "<prefix>"              # Prefix to use for policy keys in the persistence backend (no default).
  timeout: "<duration>"         # Connection timeout for the persistence backend (no default).
  etcd:
    sync: "<duration>"          # Synchronization period for an ETCD backend (no default).
    user: "<user>"              # User-id to authenticate with an ETCD backend (no default).
    password: "<password>"      # Password to authenticate with an ETCD backend (no default).
  consul:
    token: "<token>"            # Token to authenticate with a Consul backend (no default).
    namespace: "<namespace>"    # Namespace to use with a Consul backend (no default).
  postgres:
    url: "<url>"                # URL to connect and authenticate to a Postgres backend (no default).
    table: "<name>"             # Name of the table to use with a Postgres backend (no default).
    connection:
      ttl: "<duration>"         # Timeout for closing inactive Postgres connections (default "5m").
      max: <number>             # Maximum number of connections to the Postgres backend (default 100).

authentication:   # options used during the authentication stage of a user interface request.
  type: "<type>"                # Type of authentication to perform on users (default "bcrypt").

authorization:   # options used during the authorization stage of a user interface request.
  authenticate: true|false      # Flag to force authentication of the user before performing authorization (default false).

policies:   # policies used for authorizing user interface requests.
  language: "<language>"        # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
  store:
    path: "<directory>"         # Location on disk where static policy files can be found (no default).
    recurse: true|false         # Flag to indicate the given directory and its subdirectories must be search recursively for policy files (default false).

pip:   # attributes used for authorizing user interface requests.
  store:
    path: "<directory>"         # Location on disk where static attribute files can be found (no default).
    recurse: true|false         # Flag to indicate the given directory and its subdirectories must be search recursively for attribute files (default false).
  pull:   # see the section "Pull configurations" below for more information.
    configPath: "<file>"        # Location on disk where a "pull configuration" file can be found (no default).

cerbos:   # options for connecting with a cerbos PDP (sidecar) for authorizing a user interface request.
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
MANAGER_ADDRESS=<address>                       # address the service should listen on (default "0.0.0.0"; all host addresses).
MANAGER_PORT=<port>                             # port the service should listen on (default 8080).
MANAGER_TLA_CA=<certificate-file>               # CA certificate to use with the service; turns on https support (no default).
MANAGER_TLS_CERT=<certificate-file>             # TLS certificate to use with the service; turns on https support (no default).
MANAGER_TLS_KEY=<key-file>                      # TLS private key to use with the service; turns on https support (no default).
MANAGER_READ_TIMEOUT=<duration>                 # the read timeout for API requests (default "30s").
MANAGER_WRITE_TIMEOUT=<duration>                # the read timeout for API requests (default "30s").
MANAGER_IDLE_TIMEOUT=<duration>                 # the read timeout for API requests (default "300s").
MANAGER_MAX_BODY_SIZE=<size>                    # maximum allowed size of a request body (default 65536).

# options for the application-log.
MANAGER_LOG_OUTPUT=<destination>                # File or standard stream for writing the log (default "stdout").
MANAGER_LOG_FORMAT=<encoding>                   # Type of encoding for the log; supported are "text" and "json" (default "json").
MANAGER_LOG_LEVEL=<verbosity>                   # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
MANAGER_LOG_SOURCE=true|false                   # Flag to record the source location of the message in the log (default true).

# this is where the Manager stores policies, attributes, entities and relations; see the section "About persistence" below for examples.
MANAGER_PERSIST_TYPE=<type>                     # Type of persistence backend; supported are "etcd", "consul", "postgres" (no default).
MANAGER_PERSIST_ADDRESSES=<addresses>           # One or more addresses of persistence services (no default).
MANAGER_PERSIST_PREFIX=<prefix>                 # Prefix to use for policy keys in the persistence backend (no default).
MANAGER_PERSIST_TIMEOUT=<duration>              # Connection timeout for the persistence backend (no default).
MANAGER_PERSIST_ETCD_SYNC=<duration>            # Synchronization period for an ETCD backend (no default).
MANAGER_PERSIST_ETCD_USER=<user>                # User-id to authenticate with an ETCD backend (no default).
MANAGER_PERSIST_ETCD_PASSWORD=<password>        # Password to authenticate with an ETCD backend (no default).
MANAGER_PERSIST_CONSUL_TOKEN=<token>            # Token to authenticate with a Consul backend (no default).
MANAGER_PERSIST_CONSUL_NAMESPACE=<namespace>    # Namespace to use with a Consul backend (no default).
MANAGER_PERSIST_POSTGRES_URL=<url>              # URL to connect and authenticate to a Postgres backend (no default).
MANAGER_PERSIST_POSTGRES_TABLE=<name>           # Name of the table to use with a Postgres backend (no default).
MANAGER_PERSIST_POSTGRES_CONN_TTL=<duration>    # Timeout for closing inactive Postgres connections (default "5m").
MANAGER_PERSIST_POSTGRES_CONN_MAX=<number>      # Maximum number of connections to the Postgres backend (default 100).

# options used during the authentication stage of a user interface request.
MANAGER_AUTHENTICATION_TYPE=<type>              # Type of authentication to perform on users (default "bcrypt").

# options used during the authorization stage of a user interface request.
MANAGER_AUTHORIZATION_AUTHENTICATE=true|false   # Flag to force authentication of the user before performing authorization (default false).

# policies used for authorizing user interface requests.
MANAGER_POLICIES_LANGUAGE=<language>            # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
MANAGER_POLICIES_STORE=<directory>              # Location on disk where static policy files can be found (no default).
MANAGER_POLICIES_STORE_RECURSE=true|false       # Flag to indicate the given directory and its subdirectories must be search recursively for policy files (default false).

# attributes used for authorizing user interface requests.
MANAGER_MANAGER_STORE=<directory>                   # Location on disk where static attribute files can be found (no default).
MANAGER_MANAGER_STORE_RECURSE=true|false            # Flag to indicate the given directory and its subdirectories must be search recursively for attribute files (default false).
MANAGER_PULL_CONFIGS=<file>                     # Location on disk where a "pull configuration" file can be found (no default).

# options for connecting with a cerbos PDP (sidecar) for authorizing user interface requests.
MANAGER_CERBOS_ADDRESS=<address>                # Address of the Cerbos API.
MANAGER_CERBOS_ADMIN=<address>                  # Address of the Cerbos admin API.
MANAGER_CERBOS_USER=<user>                      # User-id to authenticate with the Cerbos admin API.
MANAGER_CERBOS_PSWD=<user>                      # User-id to authenticate with the Cerbos admin API.
MANAGER_CERBOS_CA=<file>                        # CA certificate to use with the Cerbos APIs.
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

# this is where the Manager stores policies, attributes, entities and relations; see the section "About persistence" below for examples.
--persist-type=<type>                     # Type of persistence backend; supported are "etcd", "consul", "postgres", "memory" (default "memory").
--persist-addresses=<addresses>           # One or more addresses of persistence services (no default).
--persist-timeout=<prefix>                # Prefix to use for policy keys in the persistence backend (no default).
--persist-timeout=<duration>              # Connection timeout for the persistence backend (no default).
--persist-etcd-sync=<duration>            # Synchronization period for an ETCD backend (no default).
--persist-etcd-user=<user>                # User-id to authenticate with an ETCD backend (no default).
--persist-etcd-password=<password>        # Password to authenticate with an ETCD backend (no default).
--persist-consul-token=<token>            # Token to authenticate with a Consul backend (no default).
--persist-consul-namespace=<namespace>    # Namespace to use with a Consul backend (no default).
--persist-postgres-url=<url>              # URL to connect and authenticate to a Postgres backend (no default).
--persist-postgres-table=<name>           # Name of the table to use with a Postgres backend (no default).
--persist-postgres-conn-ttl=<duration>    # Timeout for closing inactive Postgres connections (default "5m").
--persist-postgres-conn-max=<number>      # Maximum number of connections to the Postgres backend (default 100).

# options used during the authentication stage of a user interface request.
--authentication-type=<type>              # Type of authentication to perform on users (default "bcrypt").

# options used during the authorization stage of a user interface request.
--authorization-authenticate=true|false   # Flag to force authentication of the user before performing authorization (default false).

# policies used for authorizing user interface requests.
--policies-language=<language>            # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
--policies-store=<directory>              # Location on disk where static policy files can be found (no default).
--policies-store-recurse=true|false       # Flag to indicate the given directory and its subdirectories must be search recursively for policy files (default false).

# attributes used for authorizing user interface requests.
--pip-store=<directory>                   # Location on disk where static attribute files can be found (no default).
--pip-store-recurse=true|false            # Flag to indicate the given directory and its subdirectories must be search recursively for attribute files (default false).
--pip-pull-configs=<file>                 # Location on disk where a "pull configuration" file can be found (no default).

# options for connecting with a cerbos PDP (sidecar) for authorizing user interface requests.
--cerbos-address=<address>                # Address of the Cerbos API.
--cerbos-admin=<address>                  # Address of the Cerbos admin API.
--cerbos-user=<user>                      # User-id to authenticate with the Cerbos admin API.
--cerbos-pswd=<user>                      # User-id to authenticate with the Cerbos admin API.
--cerbos-ca=<file>                        # CA certificate to use with the Cerbos APIs.
```

### Pull configurations

A pull configuration is used to schedule pulling attributes from external PIP systems, such as IAM, HR, etc.
Retrieved attributes are cached in the Manager.
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

### About persistence

Attributes, entities and relations are persisted in a key-value store.
This is handled with the [Golang Valkeyrie library](https://github.com/kvtools/valkeyrie).

The following storage backends are currently supported:
- **Postgres**: using a custom-built Valkeyrie interface.
- **etcd**: using the standard Valkeyrie implementation.
- **Consul**: using the standard Valkeyrie implementation.
- **in-memory**: using a custom-built Valkeyrie interface (non-persistent).

If no persistence backend is configured, the **in-memory** backend will be used.
This means that when the service is restarted, all created and/or updated attributes, entities and/or relations will be gone.
For proper persistence, please configure the use of **Postgres**, **etcd** or **Consul**.

Example of a Postgres configuration:
```yaml
persist:
  type: "postgres"
  postgres:
    url: "postgres://postgres:******@postgres1:5432/open_ftv?sslmode=disable"
    table: "data"
    connection:
      ttl: "120s"
      max: 20
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

[Licensed under the EUPL](../../LICENSE.md)
