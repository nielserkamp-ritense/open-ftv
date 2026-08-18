# OpenFTV

OpenFTV is a federated authorization system: organizations administer policies and
enforce access decisions on cross-organization data requests. This glossary fixes the
vocabulary for the **management plane** — where policies are authored and the access to
that administration is itself governed.

## Language

**Component**:
A functional unit of the codebase — a directory of cohesive Go packages such as
`eam/config`, `eam/pdp/opa-embedded`, `apps/manager`, or `utilities`. Components are not
separately versioned or published; the repository is a single Go module
(see [ADR 0003](docs/adr/0003-single-go-module.md)).
_Avoid_: module (that means the one Go module of the repo).

**Management plane**:
The administration surface — the `manager` app and its UI — where authorization policies
are authored, versioned, and published. Distinct from the request path where actual
cross-organization data requests are evaluated.
_Avoid_: control plane, admin panel.

**Principal**:
The subject that acts on the management plane, identified by the `sub` of a validated
token (or a `*SYSTEM*`-style sentinel for the manager's own actions). A Principal has two
lifetimes: the **request-scoped** value the PEP derives per request, carrying the roles and
attributes the PDP evaluates, and the **stored** record the manager keeps of every
Principal it has seen, holding only display data (name, email) so past actions can be
attributed to a nameable, linkable subject. The stored record is display-facing and is
never an input to a **decision** — roles stay in the token and the PDP stays authoritative
(see [ADR 0002](docs/adr/0002-roles-as-open-data-flat-claim.md)).
_Avoid_: subject, account, identity, user (when you mean the principal specifically);
actor (that is the same thing).

**Role**:
An operator-defined label carried in the token's flat top-level `roles` claim, which
policies reference to grant capabilities. The role set is **open** — operators define
their own roles in the IdP and write policies against them. The shipped `admin` / `author`
/ `auditor` roles are example defaults, not a fixed taxonomy.
_Avoid_: group, permission, scope.

**Capability**:
Something a principal is allowed to do in the management UI (write, publish, …), derived
from a policy **decision**, not hardcoded to role names. The PDP is authoritative;
UI capabilities are affordances that should track it.
_Avoid_: permission, grant, right.

**Policy**:
A rule (Cedar, in the embedded-PDP deployment) that the PDP evaluates to permit or deny a
request. Policies are administered in the PAP; the backing store (postgres, when
configured) is the source of truth and UI edits win over the bundled seed files.
_Avoid_: rule (when you mean a Policy), permission.

**Seed policy**:
A bundled `.cedar` file loaded into an empty store at first boot to bootstrap enforcement.
Once seeded it is an ordinary, editable policy — the files are bootstrap, not an
authoritative baseline that gets re-asserted.
_Avoid_: default policy, built-in policy.

**Authorization Decision Log (ADL)**:
The record of authorization **decisions** the PDP has made — one entry per evaluated
request — kept in its own database, separate from the **policy** store. Surfaced in the
management UI as *Logboek*.
_Avoid_: audit log, authlog, decision log (spell it out on first use).

## Access-control roles (PEP / PDP / PIP / PAP)

The XACML / AuthZEN role acronyms and how OpenFTV maps them to packages are documented
in [docs/type.md](docs/type.md). In short: the **PEP** enforces, the **PDP** decides, the
**PIP** informs, the **PAP** administers.
