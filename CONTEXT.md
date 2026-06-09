# OpenFTV

OpenFTV is a federated authorization system: organizations administer policies and
enforce access decisions on cross-organization data requests. This glossary fixes the
vocabulary for the **management plane** — where policies are authored and the access to
that administration is itself governed.

## Language

**Management plane**:
The administration surface — the `manager` app and its UI — where authorization policies
are authored, versioned, and published. Distinct from the request path where actual
cross-organization data requests are evaluated.
_Avoid_: control plane, admin panel.

**Principal**:
The authenticated subject of an authorization request, identified from a validated token
as `user::<sub>`. Carries roles and other attributes the PDP evaluates.
_Avoid_: subject, account, identity, user (when you mean the principal specifically).

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

## Access-control roles (PEP / PDP / PIP / PAP)

The XACML / AuthZEN role acronyms and how OpenFTV maps them to packages are documented
in [docs/type.md](docs/type.md). In short: the **PEP** enforces, the **PDP** decides, the
**PIP** informs, the **PAP** administers.
