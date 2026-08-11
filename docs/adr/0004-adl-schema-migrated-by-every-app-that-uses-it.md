# The ADL schema is owned by its component and migrated by every app that uses it

## Executive Summary
The Authorization Decision Log (ADL) database has no single owning app. Every app that is
configured with a PostgreSQL ADL migrates that database up at startup, because:

1. There is no app that is present in every topology — the in-app deployment has no PDP at
   all, and a standalone PDP has no manager.
2. Otherwise, whether the log works depends on the startup order of two independently
   deployed apps.
3. The migrations are idempotent and concurrency-safe, so a second migrator costs nothing.

## Description
The ADL schema (`decision`, `request_type`) is embedded in the `eam/log/decisions`
component (`eam/log/decisions/migrations/postgresql/`). Until now only the PDP app ran it
(`apps/pdp/server/auth.go`, gated on `PDP_MIGRATE_*`), while the manager app only *read*
the database to serve the Logboek page. That produced
`ERROR: relation "decision" does not exist` in two ways:

- `docker/compose-mgmt-inapp.yaml` runs **no PDP service** — the manager is its own PEP with
  an embedded PDP — yet configures a PostgreSQL ADL. Nothing ever created the schema.
- In the multi-service composes the manager can start before the PDP, so Logboek stayed
  broken until the PDP happened to migrate.

So the schema belongs to the **component**, not to an app: any app configured with a
PostgreSQL ADL brings it up to date at startup. `golang-migrate` takes a Postgres advisory
lock and treats "no change" as success, so simultaneous startup of a manager and a PDP
against the same database is safe.

## Considered Options

- **The manager owns the schema and the PDP stops migrating.** Rejected: it breaks the
  standalone-PDP topology, which must keep working without a manager.
- **Neither app migrates; a separate migration job owns the schema.** Rejected: it is the
  most explicit option, but changes the operational model of every deployment for a schema
  that currently has one migration.

## Consequences

- **The manager migrates a database it never writes to.** This is deliberate, not an
  oversight — see the next point.
- **Nothing writes decisions in the in-app deployment.** The manager's embedded PDP is
  built without an ADL writer (`apps/manager/server/auth.go`, `newController`), so with no
  PDP service present Logboek loads correctly but stays empty, and
  `MANAGER_ADL_SERVICE_NAME` has no effect. Making the manager log its own
  self-authorization decisions is a separate, deferred decision.
- **ADL migration has its own configuration**, `ADL_MIGRATE_SOURCE` / `_AUTO` / `_STEPS`,
  because the ADL database is not the app's own database. `ADL_MIGRATE_SOURCE` defaults to
  the embedded scripts, and when `_AUTO`/`_STEPS` are unset they are inherited from the
  app's existing `MIGRATE_AUTO`/`MIGRATE_STEPS`, so existing deployments are fixed without
  an environment change.
- **`MIGRATE_SOURCE` is inherited by the PDP but not by the manager.** For the PDP that
  setting has only ever meant ADL scripts (it has no other database), so it must keep being
  honoured; for the manager it points at the manager's own `openftv` scripts, which must
  never be applied to the ADL database. The asymmetry lives in the two call sites, not in
  the shared helper.
- **A failed ADL migration is fatal.** A manager configured with a PostgreSQL ADL refuses
  to start if the schema cannot be brought up to date, consistent with the other fatal
  init failures in `initRoutes`. The accepted cost: an ADL database outage takes policy
  administration down with it, rather than silently serving a management plane that cannot
  produce its audit trail.
