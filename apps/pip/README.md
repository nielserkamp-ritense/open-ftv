# OpenFTV - Policy Information Point

# Welcome
This code module implements a Policy Information Point (PIP).

It supports the following interfaces:
- API endpoints to manage attributes, entities and relations; intended for user interfaces.
- API endpoints to push attributes; intended for external PIP systems, such as HR, IAM, etc.
- functionality to pull attributes from external PIPs.
- functionality to push a batch of attributes, entities and/or relations to a PDP (**TODO**).

## Building and running

See the [README](../../README.md) in the top-level folder.

Also see the notes about persistence at the end of this README.

## Docker images

The Gitlab CI/CD is set up to automatically create docker images in the
[Gitlab container registry](https://gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/container_registry/8582424).

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
svc:   # options for the main API endpoints.
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

health:   # options for the liveness & readiness API endpoints.
  host: "<address>"             # Address the service should listen on (default 'main' address).
  port: <port>                  # Port the service should listen on (default 8080).
  timeout:
    read: "<duration>"          # The read timeout for API requests (default "30s").
    write: "<duration>"         # The read timeout for API requests (default "30s").
    idle: "<duration>"          # The read timeout for API requests (default "300s").
  maxBody: <size>               # Maximum allowed size of a request body (default 65536).
  cors:
    origins: "<origins>"        # CORS origins (default "*").
    headers: "<headers>"        # CORS headers (default "*").

log:   # options for the application-log.
  output: "<destination>"       # File or standard stream for writing the log (default "stdout").
  format: "<encoding>"          # Type of encoding for the log; supported are "text" and "json" (default "json").
  level: "<verbosity>"          # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
  source: true|false            # Flag to record the source location of the message in the log (default true).

persist:   # this is where the PIP stores attributes, entities and relations; see the section "About persistence" below for examples.
  type: "<type>"                # Type of persistence backend; supported are "etcd", "consul" and "postgres" (no default).
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
    path: "<folder>"            # Location on disk where static policy files can be found (no default).
    recurse: true|false         # Flag to indicate the given folder and subfolders must be search recursively for policy files (default false).

pip:   # attributes used for authorizing user interface requests.
  store:
    path: "<folder>"         # Location on disk where static attribute files can be found (no default).
    recurse: true|false         # Flag to indicate the given folder and subfolders must be search recursively for attribute files (default false).
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
# options for the main API endpoints.
PIP_ADDRESS=<address>                       # address the service should listen on (default "0.0.0.0"; all host addresses).
PIP_PORT=<port>                             # port the service should listen on (default 8443).
PIP_TLA_CA=<certificate-file>               # CA certificate to use with the service; turns on https support (no default).
PIP_TLS_CERT=<certificate-file>             # TLS certificate to use with the service; turns on https support (no default).
PIP_TLS_KEY=<key-file>                      # TLS private key to use with the service; turns on https support (no default).
PIP_READ_TIMEOUT=<duration>                 # the read timeout for API requests (default "30s").
PIP_WRITE_TIMEOUT=<duration>                # the read timeout for API requests (default "30s").
PIP_IDLE_TIMEOUT=<duration>                 # the read timeout for API requests (default "300s").
PIP_MAX_BODY_SIZE=<size>                    # maximum allowed size of a request body (default 65536).
PIP_CORS_ORIGINS=<origins>                  # CORS origins (default "*").
PIP_CORS_HEADERS=<headers>                  # CORS headers (default "*").

# options for the liveness & readiness API endpoints.
PIP_HEALTH_HOST=<address>                   # Address the service should listen on (default 'main' address).
PIP_HEALTH_PORT=<port>                      # Port the service should listen on (default 8080).
PIP_HEALTH_READ=<duration>                  # The read timeout for API requests (default "30s").
PIP_HEALTH_WRITE=<duration>                 # The read timeout for API requests (default "30s").
PIP_HEALTH_IDLE=<duration>                  # The read timeout for API requests (default "300s").
PIP_HEALTH_MAX_BODY=<size>                  # Maximum allowed size of a request body (default 65536).
PIP_HEALTH_ORIGINS=<origins>                # CORS origins (default "*").
PIP_HEALTH_HEADERS=<headers>                # CORS headers (default "*").

# options for the application-log.
PIP_LOG_OUTPUT=<destination>                # File or standard stream for writing the log (default "stdout").
PIP_LOG_FORMAT=<encoding>                   # Type of encoding for the log; supported are "text" and "json" (default "json").
PIP_LOG_LEVEL=<verbosity>                   # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
PIP_LOG_SOURCE=true|false                   # Flag to record the source location of the message in the log (default true).

# this is where the PIP stores policies; see the section "About persistence" below for examples.
PIP_PERSIST_TYPE=<type>                     # Type of persistence backend; supported are "etcd", "consul", "postgres" (no default).
PIP_PERSIST_ADDRESSES=<addresses>           # One or more addresses of persistence services (no default).
PIP_PERSIST_PREFIX=<prefix>                 # Prefix to use for policy keys in the persistence backend (no default).
PIP_PERSIST_TIMEOUT=<duration>              # Connection timeout for the persistence backend (no default).
PIP_PERSIST_ETCD_SYNC=<duration>            # Synchronization period for an ETCD backend (no default).
PIP_PERSIST_ETCD_USER=<user>                # User-id to authenticate with an ETCD backend (no default).
PIP_PERSIST_ETCD_PASSWORD=<password>        # Password to authenticate with an ETCD backend (no default).
PIP_PERSIST_CONSUL_TOKEN=<token>            # Token to authenticate with a Consul backend (no default).
PIP_PERSIST_CONSUL_NAMESPACE=<namespace>    # Namespace to use with a Consul backend (no default).
PIP_PERSIST_POSTGRES_URL=<url>              # URL to connect and authenticate to a Postgres backend (no default).
PIP_PERSIST_POSTGRES_TABLE=<name>           # Name of the table to use with a Postgres backend (no default).
PIP_PERSIST_POSTGRES_CONN_TTL=<duration>    # Timeout for closing inactive Postgres connections (default "5m").
PIP_PERSIST_POSTGRES_CONN_MAX=<number>      # Maximum number of connections to the Postgres backend (default 100).

# options used during the authentication stage of a user interface request.
PIP_AUTHENTICATION_TYPE=<type>              # Type of authentication to perform on users (default "bcrypt").

# options used during the authorization stage of a user interface request.
PIP_AUTHORIZATION_AUTHENTICATE=true|false   # Flag to force authentication of the user before performing authorization (default false).

# policies used for authorizing user interface requests.
PIP_POLICIES_LANGUAGE=<language>            # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
PIP_POLICIES_STORE=<folder>                 # Location on disk where static policy files can be found (no default).
PIP_POLICIES_STORE_RECURSE=true|false       # Flag to indicate the given folder and subfolders must be search recursively for policy files (default false).

# attributes used for authorizing user interface requests.
PIP_PIP_STORE=<folder>                      # Location on disk where static attribute files can be found (no default).
PIP_PIP_STORE_RECURSE=true|false            # Flag to indicate the given folder and subfolders must be search recursively for attribute files (default false).
PIP_PULL_CONFIGS=<file>                     # Location on disk where a "pull configuration" file can be found (no default).

# options for connecting with a cerbos PDP (sidecar) for authorizing user interface requests.
PIP_CERBOS_ADDRESS=<address>                # Address of the Cerbos API.
PIP_CERBOS_ADMIN=<address>                  # Address of the Cerbos admin API.
PIP_CERBOS_USER=<user>                      # User-id to authenticate with the Cerbos admin API.
PIP_CERBOS_PSWD=<user>                      # User-id to authenticate with the Cerbos admin API.
PIP_CERBOS_CA=<file>                        # CA certificate to use with the Cerbos APIs.
```

### Command-line flags

The following is an exhaustive list of all possible command-line flags the service will understand.
These match the corresponding options in a configuration file.

```text
# options for the main API endpoints.
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

# options for the liveness & readiness API endpoints.
--health-host=<address>                   # address the service should listen on (default 'main' address).
--health-port=<port>                      # port the service should listen on (default 8443).
--health-read=<duration>                  # the read timeout for API requests (default "30s").
--health-write=<duration>                 # the read timeout for API requests (default "30s").
--health-idle=<duration>                  # the read timeout for API requests (default "300s").
--health-max-body=<size>                  # maximum allowed size of a request body (default 65536).
--health-origins=<origins>                # CORS origins (default "*").
--health-headers=<headers>                # CORS headers (default "*").

# options for the application-log.
--log-output=<destination>                # File or standard stream for writing the log (default "stdout").
--log-format=<encoding>                   # Type of encoding for the log; supported are "text" and "json" (default "json").
--log-level=<verbosity>                   # Verbosity level of the log; supported are "debug", "info", "warn" and "error" (default "info").
--log-source=true|false                   # Flag to record the source location of the message in the log (default true).

# this is where the PIP stores policies; see the section "About persistence" below for examples.
--persist-type=<type>                     # Type of persistence backend; supported are "etcd", "consul", "postgres" (no default).
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
--policies-store=<folder>                 # Location on disk where static policy files can be found (no default).
--policies-store-recurse=true|false       # Flag to indicate the given folder and subfolders must be search recursively for policy files (default false).

# attributes used for authorizing user interface requests.
--pip-store=<folder>                      # Location on disk where static attribute files can be found (no default).
--pip-store-recurse=true|false            # Flag to indicate the given folder and subfolders must be search recursively for attribute files (default false).
--pip-pull-configs=<file>                 # Location on disk where a "pull configuration" file can be found (no default).

# options for connecting with a cerbos PDP (sidecar) for authorizing user interface requests.
--cerbos-address=<address>                # Address of the Cerbos API.
--cerbos-admin=<address>                  # Address of the Cerbos admin API.
--cerbos-user=<user>                      # User-id to authenticate with the Cerbos admin API.
--cerbos-pswd=<user>                      # User-id to authenticate with the Cerbos admin API.
--cerbos-ca=<file>                        # CA certificate to use with the Cerbos APIs.
```

### Authentication & authorization

#### EAM inside EAM.

The PIP service has built-in EAM components. This may sound like a contradiction.
However, the policies and data used for this are stored outside the application.
Only the generic EAM code is embedded, as a proof of concept that a PDP, PAP and/or PIP do not necessarily need to be externalized.
This can save a considerable amount of resource usage in your setup.
It also significantly reduces latency between the various components.

#### Policies and data

The policies for authorization are loaded from the location indicated by the ```policies.store``` configuration parameter.
The language of these policies should match the ```policies.language``` configuration parameter,
so the embedded EAM controller knows which PDP engine to use (OPA, Cedar, Cerbos or OpenFGA).

Local data for authentication and authorization is loaded from the location indicated by the ```pip.store``` configuration parameter.
External data can be loaded with the pull configurations (see below).
This data is kept separate from the data maintained by the PIP.

The policies and local data used by the embedded EAM components are meant to be read-only,
and not to be maintained through the UI API of the PIP service.
This is to ensure separation of duties!

To modify policies, you need to restart the service with the modified policies.

To maintain data elements (such as users), you can use the pull- or push-mechanisms of the embedded PIP.
Or, alternatively, maintain a file (see below) and restart the service with the modified user-file.

#### Authentication

Authentication is configurable with the ```authentication.type``` configuration option.

Currently supported methods are:
- ```none``` = no authentication (for quick testing).
- ```bcrypt``` = bcrypt hashed password.
- ```basic``` = basic HTTP authentication (**TODO**).
- ```jwt``` = token-based authentication (**TODO**).

The authentication module requires a list of users with passwords.
This list is configured through the embedded PIP as entity-styled data.

Add a file to the configured ```pip.store``` folder as follows:
```yaml
- type: "user"
  id: "mickey"
  attributes:
    name: "Mickey Mouse"
    # bcrypt -nB mickey
    password: "....."
    roles: ["admin"]
- type: "user"
  id: "goofy"
  attributes:
    name: "Goofy"
    # bcrypt -nB goofy
    password: "....."
    roles: ["auditor"]
```

This will provide the authentication module with the users that are allowed to access the PIP service.

Note that if this list requires frequent changes,
it will save time to use the push- and/or pull-mechanism of the embedded PIP.

#### Authorization

Authorization is configured with a single parameter (```authorization.authenticate```),
and locally stored policies (```policies.store```).

The ```authorization.authenticate``` parameter indicates that a user must be authenticated first.
If this parameter is ```false``` (the default), authentication will be skipped.
Turning off authentication can be useful in test scenarios, or when it has already been performed elsewhere.

If one or more policies were found in the ```policies.store``` folder,
they will be loaded into the embedded PDP at startup and used for authorizing all UI API requests.

If no policies are found at startup, authorization is bypassed.
This is useful for quick testing, but must not be used in production setups.

#### Cerbos

Most PDP engines are written in Golang and are thus easily embedded.
The main exception is Cerbos, where the team developed the engine code in an internal folder
which is not accessible to outside projects.
This is why Cerbos needs to be configured as a sidecar.

The Cerbos admin API needs to be accessible to allow the PAP to push its policies for authorization to the PDP.
Check the [Cerbos documentation](https://docs.cerbos.dev/cerbos/latest/what-is-cerbos)
on how to install and configure it.

### Pull configurations

A pull configuration is used to schedule pulling attributes from external PIP systems, such as IAM, HR, etc.
Retrieved attributes are cached in the PIP.
The cache is refreshed after a certain amount of time according to the given schedule.

The location of this file is given by the ```pip.pull.configPath``` configuration option.
Its use is optional.

A pull configuration is a YAML encoded file with the following layout:
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

(*2) Parameters for the request URI path replace placeholders marked with two colon characters; e.g., ```:param1:```, ```:param2:```.
For this to work, the ```name``` of the parameter must match the identifier between the two colons.
Parameters for the request URI query are added as query parameters to the URI.
Parameters for the request body are structured into an object and encoded as JSON or YAML, according to the ```contentType``` option.

(*3) Options to describe the schedule or interval.
Either ```interval``` (with an optional ```initialInterval```) or ```schedule``` must be defined.
If ```interval``` is defined, the request will be executed repeatedly, with the given **interval** between calls.
If the ```schedule``` rule is defined, the request will be executed according to standard crontab rules.

(*4) see [response.go](../../eam/pip/network/response.go) for details of the mapping mechanism.

Note that pull configurations may also be encoded in JSON or TOML format.
However, for brevity, we will only include the YAML layout above, as a JSON or TOML layout can be inferred from it.

### About persistence

Attributes, entities and relations are persisted in an SQL database or key-value store.

Key-value store support is handled with the [Golang Valkeyrie library](https://github.com/kvtools/valkeyrie).
Note that key-value support is deprecated. Please use the SQL database option.

The following storage backends are currently supported:
- ```PostgreSQL```: using a custom-built Valkeyrie interface. ** RECOMMENDED **
- ```etcd```: using the standard Valkeyrie implementation. ** DEPRECATED **
- ```Consul```: using the standard Valkeyrie implementation. ** DEPRECATED **

A persistence backend must be configured explicitly; there is no default. If ```persist.type``` is
left unset or set to an unsupported value, the service fails to start with a clear error instead of
silently falling back to non-persistent, in-memory storage.

For proper persistence, please configure the use of ```PostgreSQL```.

Example of a Postgres configuration:
```yaml
persist:
  type: "PostgreSQL"
  postgres:
    url: "postgres://postgres:******@postgres1:5432/open_ftv?sslmode=disable"
    table: "data"
    connection:
      ttl: "120s"
      max: 20
```

Note that the *type* parameter is case-insensitive.

For SQL databases (such as PostgreSQL) you previously had to make sure the database and table existed using initialization scripts.
In the new version (2025/08/14), you can use the built-in database migrations (use a separate PAP or Manager service for this).

## Application log

The service writes a single application log of relevant events, so that it can aid in:
- monitoring the functioning of the service,
- debugging the root cause of problems,
- monitoring life-time events, such as displaying configuration options at startup.

Various configuration options can be used to:
- determine where the log is written (```log.output```),
- how to format the log (```log.format```),
- the verbosity level at which events will be logged (```log.level```).

Options for logging are processed first at startup.
This means the log is ready before any other options are checked and possible errors are printed in the log.
If logging options cause errors, these are printed using the standard Golang log mechanism.

## License

[Licensed under the EUPL](../../LICENSE.md)
