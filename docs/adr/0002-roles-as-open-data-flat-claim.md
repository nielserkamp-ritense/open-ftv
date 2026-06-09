# Roles are open data carried in a flat top-level claim

Roles are **open, operator-defined data**, not a fixed taxonomy. Operators define roles in
their IdP and write policies that reference them by name; the shipped `admin` / `author` /
`auditor` roles are example defaults, not the contract. OpenFTV requires the IdP to emit
roles as a **flat, top-level claim in the access token** — claim name configurable via
`OIDC_ROLES_CLAIM` (default `roles`). The PEP reads a single-level key and does no
nested-claim traversal.

We deliberately do **not** consume Keycloak's default nested `realm_access.roles` (or
`resource_access.<client>.roles`). Keeping the claim flat keeps the PEP IdP-agnostic and
the parsing trivial; the cost is that each IdP must be configured to flatten roles into the
expected claim. The shipped Keycloak realm does this with a custom `roles` protocol mapper —
operators integrating another IdP must provide the equivalent.

## Consequences

- Adding a role is a policy + IdP-config change, with no code change in OpenFTV.
- The management UI must not hardcode role names to derive capabilities; affordances should
  come from a backend capability/decision endpoint so custom roles work automatically. (The
  current `capabilities.ts` hardcodes the three example roles and is interim.)
