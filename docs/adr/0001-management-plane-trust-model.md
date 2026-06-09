# Management-plane authorization trust model

The `manager` app is **always** the authorization authority for the management plane: it
validates the OIDC token and evaluates the request against its embedded PDP in every
deployment. The gateway (Kong with the `openftv` plugin) is **optional** infrastructure
that, when present, adds perimeter enforcement on top — the manager never trusts a
gateway-forwarded identity in place of its own check. We chose always-enforce over
"gateway takes over when present" so there is a single, consistent enforcement path
regardless of topology, and no deployment runs the management API unprotected because a
gateway was assumed but absent.

## Consequences

- **OIDC is effectively required for the manager.** With no JWKS configured, bearer tokens
  yield no principal; combined with fail-closed (below) the API denies everything. OIDC is
  a setup requirement, not an optional add-on, for any non-trivial deployment.
- **An empty policy store fails CLOSED** (deny / refuse to serve), reversing the legacy
  `NoAuth()`-on-empty default that allowed all requests when no policies were loaded
  (`eam/config/authorization.go`). Seeding still populates the happy path; the failure mode
  is "locked", not "open".
- **Self-lockout is possible and accepted.** Seeded policies are ordinary, UI-editable rows
  (seed-once, UI-wins). Deleting policies until the store is empty — or a wipe/migration —
  denies everything, including the request to add a policy back; recovery is manual DB
  intervention. This is an operator responsibility, documented rather than guarded in code.

## Status

Implemented as an opt-in **secured mode**: `AUTHORIZATION_FAIL_CLOSED_ON_EMPTY` (default
`false` to preserve the legacy fail-open behavior of the shared config used by pap/pip/pdp
and the test suite). When enabled, an empty/unreadable store denies all requests and the
manager additionally requires OIDC (it refuses to start without a JWKS URL). The
`compose-mgmt-inapp.yaml` stack enables it; enabling it for the gateway stack
(`compose-mgmt-authz.yaml`) additionally requires manager-side OIDC config aligned with the
gateway's issuer and is left as follow-up.
