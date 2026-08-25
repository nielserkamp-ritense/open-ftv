# The management plane is authorized on domain objects, not on HTTP requests

The manager has always enforced its own API with its embedded PDP (ADR 0001), but the
question it asked the PDP was shaped by HTTP: the action was derived from the verb
(`PUT` → `can_create`, `POST` → `can_update`, `DELETE` → `can_delete`, anything else →
`can_read`), the resource was `service::<request URI>`, and the context carried the whole
validated JWT, every header and the URL. Policies therefore had to match on paths
(`context.http.path == "/v1/deployment"`) to say "only administrators publish", and any
operation the verb mapping did not know fell through to `can_read` — `PATCH
/v1/policy/:id/status` was authorized as a read, so an auditor could change a policy's
status. We replace that with an information model of the management plane itself, so that
policies are written in the words the product uses (SCRUM-15, decided 2026-08-25).

## Decisions

- **Resources are the domain objects, typed by their bare singular name:** `policy`, `tag`,
  `deployment`, `setting`, `attribute`, `entity`, `language`, `principal`, `decision`. A
  resource carries only the properties a policy may need — `id`, `status`, `tags`,
  `language` — and nothing more until a policy needs it. A collection read is the same type
  with id `*`, so `resource is policy` covers a list and a single object alike. Which
  objects a principal may *see* in a list is filtering, not a decision, and is out of scope.
- **Actions are generic across types, plus a few business actions:** `read`, `create`,
  `update`, `delete` on every type; `accept` (status → accepted), `revert` (status back to
  concept, withdrawing an accepted policy for editing), `deploy` (status → deployed, and
  publishing a deployment: both mean "make it live"), `restore` (an older version becomes
  the current concept). Never `read_policy`-style names per type, so "this role may read
  everything" stays a one-line policy. `tag`/`untag` are not actions: no endpoint exists,
  and a tag-scoped rule is expressed on `update` with `resource.tags`. The Cedar action
  type is `Action`, ids without the old `can_` prefix.
  *Refined during implementation (SCRUM-142):* `revert` was added because "an accepted
  policy is not edited in place" forbids `update` on it, and the life cycle's only way
  out of accepted is back to concept; without its own action that transition was
  unauthorizable for every role.
- **The subject is `user` (or `app` for API keys) with id `sub` and the single property
  `roles`.** Display and provenance claims (`preferred_username`, `email`, `iss`) reach the
  stored principal record (ADR 0004) and never the PDP. The context carries only `time`,
  `traceparent` and `tracestate`. No JWT, no URL, no headers, no body.
- **Cedar types are short names.** Cedar entity types must be identifiers, so the canonical
  Linked Data URIs of the FTV information model cannot be used inside the embedded PDP. They
  belong at an AuthZEN wire boundary, which the manager's self-check never crosses; if the
  manager ever answers AuthZEN for itself, that boundary maps `user` ↔
  `…/ftv/def/subject/user` and back (SCRUM-57).
- **The check lives in a thin service layer, one evaluation per request.** A new component,
  `eam/management`, holds one service per resource type. A middleware validates the token
  once and yields the request-scoped Principal; the service loads the object, builds the
  PARC from it, asks the PDP once, and only then runs the store operation. A missing object
  contributes its id and no properties, and the decision is returned before the not-found:
  403 before 404, so an unauthorized caller cannot probe for ids.
- **The management-plane PARC is built in `eam/authorization`, not in `eam/pep`.** The PEP
  also serves the gateway plugin, where `context.http` and the JWT are legitimate policy
  input for cross-organization requests. The standalone PAP and PIP mount the same handlers
  as the manager and therefore inherit the lean model.
- **The role matrix of SCRUM-16 is the target for the manager's shipped example policies:**
  `author` (functioneel beheerder) edits, accepts and deploys rules and PIP data; `admin`
  (systeembeheerder) manages beslispunten and settings; everyone reads, the ADL included.
  `admin` is not a superuser. Claim values stay `admin`, `author`, `auditor` (ADR 0002); the
  `user` role is dropped from the manager's examples because `auditor` already means "read
  everything". The **standalone PAP and PIP are the exception**: they have no OIDC and their
  locally seeded users carry the roles `admin` and `user`, so their example policies keep
  those semantics (`admin` everything, `user` and API keys read) in the new model. The
  matrix is the manager's; a standalone PAP is an operator tool, not the management plane.
  Locally seeded users may carry `roles` as one string rather than a set, so the example
  policies accept both shapes (`roles == "x" || roles.contains("x")`).
- **Roles are scoped to a beslispunt by a scoped role string in the existing flat claim**
  (`author:laadpalen`), with one policy per beslispunt. ADR 0002 is unchanged; adding a
  beslispunt-scoped role is IdP configuration plus a policy. A structured claim or a
  manager-held assignment (SCRUM-18) were considered and deferred: the first needs an IdP
  mapper per deployment, the second persists roles in the manager against ADR 0004.

## Rollout

The handlers move to the service layer one resource type at a time, each its own change:
policies, deployments, tags and settings, attributes and entities, then languages and ADL
entries. Until the last one has moved, two models coexist in the policy store, so the
old-model example policies (`resource is service`, `name::"can_*"`) stay in the seed
directories and each step adds the new-model rules for its type. The final step removes
the old rules and the verb-derived fallback from the management path.

Seeding is therefore decided **per file**, no longer per store: a seed file whose id is
absent is inserted, a present row is left alone, and a row that was seeded once and deleted
by the operator is not brought back (its audit trail, which survives deletion since
ADR 0005, is the ledger). A store seeded by an earlier release receives the rules a later
release adds; "UI wins" still holds for every row the operator has touched. A store without
an audit trail (memory, etcd, consul) cannot tell "never seeded" from "seeded and deleted"
and keeps the old rule: it is seeded only while empty. Postgres is the supported store for
the management plane, so that limitation costs nothing in practice.

## Consequences

- Policies name the product's objects and operations; no policy needs to know a URL. The
  `PATCH` hole closes because a status change is `accept`/`deploy`, never a read.
- Old-model rows already in a store keep existing after the rollout completes. They match
  nothing and are safe to delete; the manager does not delete them because the operator may
  have edited them.
- `admin` loses rule editing and deployment relative to the previous example policies. The
  UI's interim capability flags are flipped to the same matrix in the deployments step;
  their replacement by PDP evaluations is SCRUM-16 / SCRUM-98.
- The standalone PAP and PIP move to the new model with the manager, without a change of
  their own, because they mount the same handlers; their example policies are rewritten in
  the same steps, keeping their own role semantics (see above).
- A per-beslispunt policy is written per beslispunt, because Cedar cannot split a role
  string. If that becomes a burden, the structured-claim option is the next candidate, and
  it can be adopted per deployment without touching the model above.
- The manager still does not log its own decisions to the ADL (deferred in ADR 0004). With
  meaningful requests that gap is now visible; it is SCRUM-148.
