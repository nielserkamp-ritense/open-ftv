# OpenFTV - Manager (PAP + PIP)

# Welcome

This code module implements a Policy Information Point (PIP) and Policy Administration Point (PAP) in a single service.

It supports the following interfaces:
- API endpoints to manage policies; intended for user interfaces.
- API endpoints to manage attributes, entities and relations; intended for user interfaces.
- API endpoints to push attributes; intended for external PIP systems, such as HR, IAM, etc.
- functionality to pull attributes from external PIPs.
- API endpoint to retrieve a bundle with policies, design-time attributes, entities and/or relations; intended for PDPs.
- functionality to push a bundle with policies, design-time attributes, entities and/or relations to a PDP.
- functionality to push a bundle with policies, design-time attributes, entities and/or relations to a git repository (**TODO**).

## Building and running

See the [README](../../README.md) in the top-level folder.

Also see the notes about persistence at the end of this README.

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
svc:   # options for the main API endpoints.
  host: "<address>"             # Address the service should listen on (default "0.0.0.0"; all host addresses).
  port: <port>                  # Port the service should listen on (default 8443).
  tls:
    ca: "<certificate-file>"    # CA certificate to use with the service; turns on https support and mTLS (no default).
    cert: "<certificate-file>"  # TLS certificate to use with the service; turns on https support (no default).
    key: "<key-file>"           # TLS private key to use with the service; turns on https support (no default).
  timeout:
    read: "<duration>"          # The read timeout for API requests (default "30s").
    write: "<duration>"         # The read timeout for API requests (default "30s").
    idle: "<duration>"          # The read timeout for API requests (default "300s").
  maxBody: <size>               # Maximum allowed size of a request body (default 65536).
  cors:
    origins: "<origins>"        # CORS origins (default "*").
    headers: "<headers>"        # CORS headers (default "*").

internal:   # options for the internal API endpoints (bundle retrieval with PDP).
  host: "<address>"             # Address the service should listen on (default 'main' address').
  port: <port>                  # Port the service should listen on (default 9443).
  tls:
    ca: "<certificate-file>"    # CA certificate to use with the service; turns on https support and mTLS (no default).
    cert: "<certificate-file>"  # TLS certificate to use with the service; turns on https support (no default).
    key: "<key-file>"           # TLS private key to use with the service; turns on https support (no default).
  timeout:
    read: "<duration>"          # The read timeout for API requests (default "30s").
    write: "<duration>"         # The read timeout for API requests (default "30s").
    idle: "<duration>"          # The read timeout for API requests (default "300s").
  maxBody: <size>               # Maximum allowed size of a request body (default 65536).
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

persist:   # this is where the Manager stores policies, attributes, entities and relations; see the section "About persistence" below for examples.
  type: "<type>"                # Type of persistence backend; supported are "etcd", "consul", "postgres" (no default).
  addresses: "<addresses>"      # One or more addresses of persistence services (no default).
  base: "<prefix>"              # Prefix to use for policy keys in the persistence backend (no default).
  timeout: "<duration>"         # Connection timeout for the persistence backend (no default).
  etcd:                         # ** DEPRECATED **
    sync: "<duration>"          # Synchronization period for an ETCD backend (no default).
    user: "<user>"              # User-id to authenticate with an ETCD backend (no default).
    password: "<password>"      # Password to authenticate with an ETCD backend (no default).
  consul:                       # ** DEPRECATED **
    token: "<token>"            # Token to authenticate with a Consul backend (no default).
    namespace: "<namespace>"    # Namespace to use with a Consul backend (no default).
  postgres:
    url: "<url>"                # URL to connect and authenticate to a Postgres backend (no default).
    table: "<name>"             # Name of the table to use with a Postgres backend (no default).
    connection:
      ttl: "<duration>"         # Timeout for closing inactive Postgres connections (default "5m").
      max: <number>             # Maximum number of connections to the Postgres backend (default 100).

migrate:    # options for migrating a persistence database.
  source: "<source>"            # Source of the migration scripts. Use "*EMBED*" to use the default embedded migration scripts.
  auto: true|false              # A `true` value turns on migration, and attempts to migrate up to the highest level. Mutually exclusive with the `steps` parameter.
  steps: <number>               # Number of steps to migrate. A negative number means to migrate down, Mutually exclusive with the `auto` parameter.
  exitAfter: true|false         # A `true` value shuts down the app after the migration.

authentication:   # options used during the authentication stage of a user interface request.
  type: "<type>"                # Type of authentication to perform on users (default "bcrypt").

authorization:   # options used during the authorization stage of a user interface request.
  authenticate: true|false      # Flag to force authentication of the user before performing authorization (default false).

policies:   # policies for authorizing user interface requests.
  language: "<language>"        # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS", "OPENFGA" (default "CEDAR").
  store: 
    path: "<folder>"            # Location on disk where static policy files can be found (no default).
    recurse: true|false         # Flag to indicate the given folder and subfolders must be search recursively for policy files (default false).

pip:   # attributes for authorizing user interface requests.
  store:
    path: "<folder>"            # Location on disk where static attribute files can be found (no default).
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

bundles:   # options for bundle management (see below).
  path: "<folder>"              # Location on disk where bundle configurations can be found.
  recurse: true|false           # Flag to indicate the given folder and subfolders must be search recursively for bundle files (default false).
  sendTimeout: "<duration>"     # The read timeout for sending bundles to a PDP (default "1m").
  workers: <number>             # Maximum number of worker threads for sending bundles (default #cpu).
  stageDelay: "<duration>"      # Forced delay between bundle deployment stages (default "0s").
```

### Environment variables

The following is an exhaustive list of all possible environment variables the service will understand.
These match the corresponding options in a configuration file.

```text
# options for the main API endpoints.
MANAGER_HOST=<address>                          # Address the service should listen on (default "0.0.0.0"; all host addresses).
MANAGER_PORT=<port>                             # Port the service should listen on (default 8443).
MANAGER_TLA_CA=<certificate-file>               # CA certificate to use with the service; turns on https support (no default).
MANAGER_TLS_CERT=<certificate-file>             # TLS certificate to use with the service; turns on https support (no default).
MANAGER_TLS_KEY=<key-file>                      # TLS private key to use with the service; turns on https support (no default).
MANAGER_READ_TIMEOUT=<duration>                 # The read timeout for API requests (default "30s").
MANAGER_WRITE_TIMEOUT=<duration>                # The read timeout for API requests (default "30s").
MANAGER_IDLE_TIMEOUT=<duration>                 # The read timeout for API requests (default "300s").
MANAGER_MAX_BODY_SIZE=<size>                    # Maximum allowed size of a request body (default 65536).
MANAGER_CORS_ORIGINS=<origins>                  # CORS origins (default "*").
MANAGER_CORS_HEADERS=<headers>                  # CORS headers (default "*").

# options for the internal API endpoints (bundle retrieval with PDP).
MANAGER_INTERNAL_HOST=<address>                 # Address the service should listen on (default 'main' address).
MANAGER_INTERNAL_PORT=<port>                    # Port the service should listen on (default 9443).
MANAGER_INTERNAL_CA=<certificate-file>          # CA certificate to use with the service; turns on https support (no default).
MANAGER_INTERNAL_CERT=<certificate-file>        # TLS certificate to use with the service; turns on https support (no default).
MANAGER_INTERNAL_KEY=<key-file>                 # TLS private key to use with the service; turns on https support (no default).
MANAGER_INTERNAL_READ=<duration>                # The read timeout for API requests (default "30s").
MANAGER_INTERNAL_WRITE=<duration>               # The read timeout for API requests (default "30s").
MANAGER_INTERNAL_IDLE=<duration>                # The read timeout for API requests (default "300s").
MANAGER_INTERNAL_MAX_BODY=<size>                # Maximum allowed size of a request body (default 65536).
MANAGER_INTERNAL_ORIGINS=<origins>              # CORS origins (default "*").
MANAGER_INTERNAL_HEADERS=<headers>              # CORS headers (default "*").

# options for the liveness & readiness API endpoints.
MANAGER_HEALTH_HOST=<address>                   # Address the service should listen on (default 'main' address).
MANAGER_HEALTH_PORT=<port>                      # Port the service should listen on (default 8080).
MANAGER_HEALTH_READ=<duration>                  # The read timeout for API requests (default "30s").
MANAGER_HEALTH_WRITE=<duration>                 # The read timeout for API requests (default "30s").
MANAGER_HEALTH_IDLE=<duration>                  # The read timeout for API requests (default "300s").
MANAGER_HEALTH_MAX_BODY=<size>                  # Maximum allowed size of a request body (default 65536).
MANAGER_HEALTH_ORIGINS=<origins>                # CORS origins (default "*").
MANAGER_HEALTH_HEADERS=<headers>                # CORS headers (default "*").

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
MANAGER_PERSIST_ETCD_SYNC=<duration>            # ** DEPRECATED ** Synchronization period for an ETCD backend (no default).
MANAGER_PERSIST_ETCD_USER=<user>                # ** DEPRECATED ** User-id to authenticate with an ETCD backend (no default).
MANAGER_PERSIST_ETCD_PASSWORD=<password>        # ** DEPRECATED ** Password to authenticate with an ETCD backend (no default).
MANAGER_PERSIST_CONSUL_TOKEN=<token>            # ** DEPRECATED ** Token to authenticate with a Consul backend (no default).
MANAGER_PERSIST_CONSUL_NAMESPACE=<namespace>    # ** DEPRECATED ** Namespace to use with a Consul backend (no default).
MANAGER_PERSIST_POSTGRES_URL=<url>              # URL to connect and authenticate to a Postgres backend (no default).
MANAGER_PERSIST_POSTGRES_TABLE=<name>           # Name of the table to use with a Postgres backend (no default).
MANAGER_PERSIST_POSTGRES_CONN_TTL=<duration>    # Timeout for closing inactive Postgres connections (default "5m").
MANAGER_PERSIST_POSTGRES_CONN_MAX=<number>      # Maximum number of connections to the Postgres backend (default 100).

# options for migration of a persistence database.
MANAGER_MIGRATE_SOURCE="<source>"               # Source of the migration scripts. Use "*EMBED*" to use the default embedded migration scripts.
MANAGER_MIGRATE_AUTO=true|false                 # A `true` value turns on migration, and attempts to migrate up to the highest level. Mutually exclusive with the `steps` parameter.
MANAGER_MIGRATE_STEPS=<number>                  # Number of steps to migrate. A negative number means to migrate down, Mutually exclusive with the `auto` parameter.
MANAGER_MIGRATE_AND_EXIT=true|false             # A `true` value shuts down the app after the migration.

# options used during the authentication stage of a user interface request.
MANAGER_AUTHENTICATION_TYPE=<type>              # Type of authentication to perform on users (default "bcrypt").

# options used during the authorization stage of a user interface request.
MANAGER_AUTHORIZATION_AUTHENTICATE=true|false   # Flag to force authentication of the user before performing authorization (default false).

# policies for authorizing user interface requests.
MANAGER_POLICIES_LANGUAGE=<language>            # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
MANAGER_POLICIES_STORE=<folder>                 # Location on disk where static policy files can be found (no default).
MANAGER_POLICIES_STORE_RECURSE=true|false       # Flag to indicate the given folder and subfolders must be search recursively for policy files (default false).

# attributes for authorizing user interface requests.
MANAGER_PIP_STORE=<folder>                      # Location on disk where static attribute files can be found (no default).
MANAGER_PIP_STORE_RECURSE=true|false            # Flag to indicate the given folder and subfolders must be search recursively for attribute files (default false).
MANAGER_PULL_CONFIGS=<file>                     # Location on disk where a "pull configuration" file can be found (no default).

# options for connecting with a cerbos PDP (sidecar) for authorizing user interface requests.
MANAGER_CERBOS_ADDRESS=<address>                # Address of the Cerbos API.
MANAGER_CERBOS_ADMIN=<address>                  # Address of the Cerbos admin API.
MANAGER_CERBOS_USER=<user>                      # User-id to authenticate with the Cerbos admin API.
MANAGER_CERBOS_PSWD=<user>                      # User-id to authenticate with the Cerbos admin API.
MANAGER_CERBOS_CA=<file>                        # CA certificate to use with the Cerbos APIs.

# options for bundle management (see below).
MANAGER_BUNDLE_CONFIGS=<folder>                 # Location on disk where bundle configurations can be found.
MANAGER_BUNDLE_RECURSE=true|false               # Flag to indicate the given folder and subfolders must be search recursively for bundle files (default false).
MANAGER_BUNDLE_SEND_TIMEOUT=<duration>          # The read timeout for sending bundles to a PDP (default "1m").
MANAGER_BUNDLE_WORKERS=<number>                 # Maximum number of worker threads for sending bundles (default #cpu).
MANAGER_BUNDLE_STAGE_DELAY=<duration>           # Forced delay between bundle deployment stages (default "0s").
```

### Command-line flags

The following is an exhaustive list of all possible command-line flags the service will understand.
These match the corresponding options in a configuration file.

```text
# options for the main API endpoints.
--host=<address>                          # address the service should listen on (default "0.0.0.0"; all host addresses).
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

# options for the internal API endpoints (bundle retrieval with PDP).
--internal-host=<address>                 # address the service should listen on (default 'main' address).
--internal-port=<port>                    # port the service should listen on (default 8443).
--internal-ca=<certificate-file>          # CA certificate to use with the service; turns on https support (no default).
--internal-cert=<certificate-file>        # TLS certificate to use with the service; turns on https support (no default).
--internal-key=<key-file>                 # TLS private key to use with the service; turns on https support (no default).
--internal-read=<duration>                # the read timeout for API requests (default "30s").
--internal-write=<duration>               # the read timeout for API requests (default "30s").
--internal-idle=<duration>                # the read timeout for API requests (default "300s").
--internal-max-body=<size>                # maximum allowed size of a request body (default 65536).
--internal-origins=<origins>              # CORS origins (default "*").
--internal-headers=<headers>              # CORS headers (default "*").

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

# this is where the Manager stores policies, attributes, entities and relations; see the section "About persistence" below for examples.
--persist-type=<type>                     # Type of persistence backend; supported are "etcd", "consul", "postgres", "memory" (default "memory").
--persist-addresses=<addresses>           # One or more addresses of persistence services (no default).
--persist-timeout=<prefix>                # Prefix to use for policy keys in the persistence backend (no default).
--persist-timeout=<duration>              # Connection timeout for the persistence backend (no default).
--persist-etcd-sync=<duration>            # ** DEPRECATED ** Synchronization period for an ETCD backend (no default).
--persist-etcd-user=<user>                # ** DEPRECATED ** User-id to authenticate with an ETCD backend (no default).
--persist-etcd-password=<password>        # ** DEPRECATED ** Password to authenticate with an ETCD backend (no default).
--persist-consul-token=<token>            # ** DEPRECATED ** Token to authenticate with a Consul backend (no default).
--persist-consul-namespace=<namespace>    # ** DEPRECATED ** Namespace to use with a Consul backend (no default).
--persist-postgres-url=<url>              # URL to connect and authenticate to a Postgres backend (no default).
--persist-postgres-table=<name>           # Name of the table to use with a Postgres backend (no default).
--persist-postgres-conn-ttl=<duration>    # Timeout for closing inactive Postgres connections (default "5m").
--persist-postgres-conn-max=<number>      # Maximum number of connections to the Postgres backend (default 100).

# options for migration of a persistence database.
--migrate-source="<source>"               # Source of the migration scripts. Use "*EMBED*" to use the default embedded migration scripts.
--migrate-auto=true|false                 # A `true` value turns on migration, and attempts to migrate up to the highest level. Mutually exclusive with the `steps` parameter.
--migrate-steps=<number>                  # Number of steps to migrate. A negative number means to migrate down, Mutually exclusive with the `auto` parameter.
--migrate-and-exit=true|false             # A `true` value shuts down the app after the migration.

# options used during the authentication stage of a user interface request.
--authentication-type=<type>              # Type of authentication to perform on users (default "bcrypt").

# options used during the authorization stage of a user interface request.
--authorization-authenticate=true|false   # Flag to force authentication of the user before performing authorization (default false).

# policies for authorizing user interface requests.
--policies-language=<language>            # Default policy language for policies; supported are "OPA", "CEDAR", "CERBOS" & "OPENFGA" (default "CEDAR").
--policies-store=<folder>                 # Location on disk where static policy files can be found (no default).
--policies-store-recurse=true|false       # Flag to indicate the given folder and subfolders must be search recursively for policy files (default false).

# attributes for authorizing user interface requests.
--pip-store=<folder>                      # Location on disk where static attribute files can be found (no default).
--pip-store-recurse=true|false            # Flag to indicate the given folder and subfolders must be search recursively for attribute files (default false).
--pip-pull-configs=<file>                 # Location on disk where a "pull configuration" file can be found (no default).

# options for connecting with a cerbos PDP (sidecar) for authorizing user interface requests.
--cerbos-address=<address>                # Address of the Cerbos API.
--cerbos-admin=<address>                  # Address of the Cerbos admin API.
--cerbos-user=<user>                      # User-id to authenticate with the Cerbos admin API.
--cerbos-pswd=<user>                      # User-id to authenticate with the Cerbos admin API.
--cerbos-ca=<file>                        # CA certificate to use with the Cerbos APIs.

# options for bundle management (see below).
--bundle-configs=<folder>                 # Location on disk where bundle configurations can be found.
--bundle-recurse=true|false               # Flag to indicate the given folder and subfolders must be search recursively for bundle files (default false).
--bundle-send-timeout=<duration>          # The read timeout for sending bundles to a PDP (default "1m").
--bundle-workers=<number>                 # Maximum number of worker threads for sending bundles (default #cpu).
--bundle-stage-delay=<duration>           # Forced delay between bundle deployment stages (default "0s").
```

### Authentication & authorization

#### EAM inside EAM.

The Manager service has built-in EAM components. This may sound like a contradiction.
However, the policies and data used for this are stored outside the application.
Only the generic EAM code is embedded, as a proof of concept that a PDP, PAP and/or PIP do not necessarily need to be externalized.
This can save a considerable amount of resource usage in your setup.
It also significantly reduces latency between the various components.

#### Policies and data

The policies for authorization are loaded from the location indicated by the ```policies.store``` configuration parameter.
These policies are kept separate from the policies maintained by the Manager service.
The language of these policies should match the ```policies.language``` configuration parameter,
so the embedded EAM controller knows which PDP engine to use (OPA, Cedar, Cerbos or OpenFGA).

Local data for authentication and authorization is loaded from the location indicated by the ```pip.store``` configuration parameter.
This data is kept separate from the data maintained by the Manager service.
External data can be loaded with the pull configurations (see below).

The policies and local data used by the embedded EAM components are meant to be read-only,
and not to be maintained through the UI API of the Manager service.
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

This will provide the authentication module with the users that are allowed to access the Manager service.

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

### Tags

Tags are used to annotate policies, but also attributes, entities and relations.

These annotation are used by the bundle management system to select the appropriate policies and data.

Tags are currently stored in config files, in either YAML or JSON format:
```yaml
- id: "<identifier>"            # unique identifier of a tag.
  name: "<name>"                # human-readable, descriptive name of a tag (used in drop-down menus).
  description: "<description>"  # optional description of a tag.
```

Each file can contain one or more tags.
The full set of tags is available from the ```/v1/tags``` endpoint.

The tag definitions in files should not contain audit details.
This is meant for a future version where tags are maintained through the API.

### Pull configurations

A pull configuration is used to schedule pulling attributes from external PIP systems, such as IAM, HR, etc.
Retrieved attributes are cached in the Manager service.
The cache is refreshed after a certain amount of time according to the given schedule.

The location of this file is governed by the ```pip.pull.configPath``` option in the general configuration.
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
Either ```interval``` (with an optional ```initialInterval```) or ```schedule```* must be defined.
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
- ```PostgreSQL```: using the pgpool & pgx libraries. ** RECOMMENDED **
- ```etcd```: using the standard Valkeyrie implementation. ** DEPRECATED **
- ```Consul```: using the standard Valkeyrie implementation. ** DEPRECATED **
- ```in-memory```: using a custom-built Valkeyrie interface (non-persistent, for caching or testing only).

If no persistence backend is configured, the ```in-memory``` backend will be used.
This means that when the service is restarted, all created and/or updated policies will be gone.

For proper persistence, please configure the use of ```PostgreSQL```.

Example of a PostgreSQL configuration:
```yaml
persist:
  type: "PostgreSQL"
  postgres:
    url: "postgres://postgres:******@postgres1:5432/open_ftv?sslmode=disable"
    table: "ftv"
    connection:
      ttl: "120s"
      max: 20
```

Note that the *type* parameter is case-insensitive.

For SQL databases (such as PostgreSQL) you previously had to make sure the database and table existed using initialization scripts.
In the new version (2025/08/14), you can use the built-in database migrations (see below).

### Database migrations

When using an SQL database for persistence (such as PostgreSQL),
the Manager service allows you to perform database migrations, either automatically or manually.

The target database must have a valid URL in the persistence configuration.
ALl other parameters are defined in the migration configuration.

The ```source``` defines where the migration scripts are located.
THis should be a disk folder or an embedded file system.

The ```auto``` and ```steps``` parameter indicate how to perform the migration.
- ```auto``` takes precedence; the value ```true``` indicates the migration must be performed to the highest possible level.
- ```steps``` can be used to manually fine-tune the migration. A positive number indicates the number of levels to migrate upwards.
  A negative number indicates the number of levels to migrate downwards.

If ```auto == false && steps == 0```, no migration takes place.

The ```exitAfter``` parameter can be set to ```true``` (default is ```false```) to make the app shut down after the migration.
This can be useful if you want to use an initialization container for just the migrations.

### Bundle management

The service can be configured to manage deployment bundles.

A bundle consists of all policies, attributes, entities and/or relations needed for a specific PDP or set of PDPs.

Check the following files and folders for implementation details: 
- *oas/bundles*: Open-API spec.
- *eam/config/bundle.go: bundle top-level configuration.
- *eam/bundles*: bundle configuration and management.
- *eam/handlers/bundles.go*: bundle UI API handlers (PAP).
- *eam/handlers/bundle_receiver.go*: bundle receiver API handler (PDP).

#### Deployment stages

The deployment of bundles involves the following stages:
- ```Creating```: determine the next version number.
- ```Gathering```: gathering all policies, attributes, entities and/or relations needed.
- ```Merging```: merge the gathered elements into a git repository (e.g., change and history management).
- ```Bundling```: assembling bundles.
- ```Sending```: preparing bundles (compression) and sending to configured PDPs.

Once all processing is complete, the status of the bundle is marked as *Completed*.
If, at any stage, an unrecoverable error occurs, the bundle is marked as *Failed* with an appropriate error message.

#### Annotation tags
The selection of which policies and data go into a specific bundle is decided by the tags an element is annotated with.

Each bundle is annotated with one or more tags.
Any element annotated with one of these tags will be included in the bundle.
Note that only a single tag needs to match.

#### Bundle configuration

A bundle configuration is a file in YAML or JSON format.

An example of a bundle configuration:
```yaml
---
id: "<identifier"                # **required** A unique identifier for the bundle.
language: "<policy language>"    # **required** The code or name of the policy language for the PDP ("REGO", "CEDAR", "CERBOS", "OPENFGA").
policies: true|false             # Flag to indicate the bundle should (or shouldn't) include matching policies.
data: true|false                 # Flag to indicate the bundle should (or shouldn't) include matching attributes, entities and/or relations.
version: true|false              # Flag to indicate the bundle should (or shouldn't) include the version number of the deployment.
tags: ["<tag>", ...]             # **required** Tags to select the elements for inclusion in the bundle.
targets: []                      # **required** A list of one or more target PDPs.
```

Example of a target PDP configuration:
```yaml
  - uri: "<uri>"                 # **required** URI on which the target PDP is listening for bundles.
    ca: "<certificate-file>"     # CA certificate to use with the send-request; turns on https.
    cert: "<certificate-file>"   # TLS certificate to use with the send-request; turns on https.
    key: "<key-file>"            # TLS private key to use with the send-request; turns on https.
    apiKey: "<api-key>"          # API key to include in the request.
    headers: ["<header1>", ...]  # Headers to include in the request.
    compress: "<type>"           # Type of compression used for the bundle ("GZIP", "BZIP2"; default "GZIP").
```

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
