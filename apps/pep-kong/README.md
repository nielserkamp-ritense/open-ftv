# OpenFTV - PEP for Kong

The Policy Enforcement Point (PEP) for the open-source Kong gateway, as a Kong Go plugin called `openftv`. It
authorizes each request through an AuthZEN compatible PDP, following the AuthZEN API gateway profile, and can also have
the PDP evaluate the upstream response.

The plugin is configured per service or route in Kong's declarative configuration; see the `openftv` plugin entries in
`demo/*/kong/conf.yaml.template` and the `Config` fields in `cmd/openftv/config.go`.

## Running

- **Locally:** docker compose builds `Dockerfile`, a Kong image with the plugin at `/usr/local/bin/openftv`.
- **Simulation:** CI builds the `gw-kong` image with ko on a Kong base image, with the plugin at `/ko-app/openftv`.

Kong names a Go plugin after its executable, and ko names the executable after the folder holding `main.go`. That
folder must therefore stay `cmd/openftv`, or Kong no longer finds the `openftv` plugin.
