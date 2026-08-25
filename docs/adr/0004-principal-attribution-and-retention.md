# Principal attribution, retention, and the upsert that backs them

The management plane stores *who* created or last changed every object as a plain string in
`created_by` / `updated_by`, and every audit row in `*_audit.user_id`. Those strings used to
be two different things — a display name in the first pair, a `sub` in the second — so the UI
could either name a person or link to them, never both. We now store the **Principal id** in
all of them and keep a `principal` table holding the display data (name, email) that turns an
id back into a nameable, linkable subject.

## Decisions

- **`principal.id` is the attribution string itself**, not a generated key: a `sub` for users, a
  `*SYSTEM*`-style sentinel for the manager's own actions, and a display name for rows carried
  over from before this change. That is what makes the change possible without rewriting a
  single attribution column. There is no separate `subject` column, because it would duplicate
  the primary key exactly; `issuer` is kept as provenance for a future second IdP.
- **`created_by` / `updated_by` are foreign keys to `principal(id)`, `ON DELETE RESTRICT`**, on
  all nine tables that carry them (17 constraints). Attribution is an active relationship, so
  the database enforces that the referenced subject exists. `RESTRICT` is what implements
  "principals are never deleted"; retirement is `active = false`, because a retired owner must
  still be visible as the owner.
- **`*_audit.user_id` stays an unconstrained plain string.** An audit write must never fail
  because a principal row is missing, and audit history must survive deprovisioning.
- **Names are resolved by id lookup at the serialisation boundary, falling back to the raw id.**
  `eam/principals` owns the table and exposes a batched resolver; the components that own
  policies, entities, settings and bundles do not join to a table they do not own.
- **The upsert runs after a permit, is required for writes and best-effort for reads.** Every
  permitted request upserts the caller's Principal. On a safe method (`GET`/`HEAD`/`OPTIONS`) a
  failure is logged and ignored; on any other method it fails the request before the handler
  runs. The rule the ticket set out to protect is that a degraded database must never lock
  everyone *out of* the management UI — that is about read access. A database too degraded to
  accept the upsert would have rejected the write anyway, so blocking there costs no
  availability and turns a foreign-key violation deep inside a handler into an early,
  comprehensible error. Gating on the decision means a token that is valid but permitted
  nothing leaves no personal data behind.
- **`principal` never gets a `roles` column.** Roles stay in the token and the PDP stays
  authoritative; a persisted role set would be a second, silently diverging authorization model
  (see [ADR 0002](0002-roles-as-open-data-flat-claim.md)).

## Consequences

- Attribution values written before this change are display names, not ids. Migration `00014`
  inserts each distinct legacy value as a `kind='legacy'` principal so the constraints validate
  without rewriting the attribution columns. Those rows are keyed by a name and are shown but
  not linked; the same human gets a second, `sub`-keyed row on next login.
- With 17 `ON DELETE RESTRICT` constraints a principal is referenced from up to 17 places, so
  there is no path to erasing a person's data without first rewriting every attribution column.
  An AVG erasure request is therefore a deliberate operational procedure, not a `DELETE`.
- `active` ships with no writer. Nothing deactivates a principal yet, because the manager has no
  deprovisioning signal from the IdP — it only ever learns of a user by seeing one succeed. The
  column exists because `ON DELETE RESTRICT` leaves it as the only retirement mechanism; the
  signal that sets it is follow-up work.
- In `noAuth` / fail-open mode (ADR 0001) every authenticated caller is permitted, so gating the
  upsert on the decision stores everyone. "Only principals allowed to act are stored" is a
  guarantee of secured mode, not of every deployment.
- `identity.Principal.DisplayName()` is removed. Its only callers were the write sites that now
  store `user.ID`, and keeping it would leave a method whose whole purpose is to put a name where
  an id belongs.
