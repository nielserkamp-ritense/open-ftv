# Contributing

How we work on OpenFTV, so the code and history read as if one person wrote them. For setup and running the
project, see the [README](README.md).

## Writing

Write in plain English and get to the point. This goes for commit messages, merge requests, docs and review
comments.

## Commits

### Format

Commits follow [Conventional Commits](https://www.conventionalcommits.org), e.g. `feat: add a Policy editor` or
`fix: answer a rejected bearer token with 401`. The version bump and the release notes are derived from them, and
the `Lint commit messages` job checks every commit on a branch. Check your branch locally with
`npx --package @commitlint/cli --package @commitlint/config-conventional commitlint --from origin/main`.

- `feat` and `fix` end up in the release notes, as do `perf`, `refactor`, `docs`, `build`, `ci` and `revert`. The
  `Revert "..."` message of `git revert` and GitLab's Revert button counts as a `revert`.
- `chore`, `style` and `test` are left out, and a range with only those has nothing to release. A breaking one is
  listed under *Other*.
- Mark a breaking change with `!` (`feat!: ...`) or with a `BREAKING CHANGE:` footer. Both bump the *minor* while
  pre-1.0. A footer needs a blank line above it, which `git commit -m "feat: ..." -m "BREAKING CHANGE: ..."` gives
  you. The linter rejects a footer without that blank line, because the release notes would otherwise drop the
  commit or miss that it is breaking.
- Keep the subject lowercase. `feat: Add a thing` is rejected.

### Subject

Say what changes, in the imperative, as the user or client sees it. The subject ends up in the release notes.

- Good: `fix: answer a failed PDP evaluation with 500 instead of 200`
- Bad: `fix: set status code in handler`

### One concern per commit

- Each commit does one thing: a refactor, a test, a fix. Don't mix them.
- Put preparation first: the refactor or test that makes the change possible comes before the change itself.
- Every commit builds and passes its tests.

### Body: the reasoning

We keep comments in the code to a minimum, so the commit body is where the *why* lives. Say what was wrong or
missing and why this is the fix. Leave the body out only when the subject already says it all.

```
fix: answer a failed PDP evaluation with 500 instead of 200

The error reply reused the 200 set once the request passed its checks.
On /evaluations without an evaluations array, a batch client then got a
200 without evaluations and could read that as all allowed.
```

When a change follows an external spec, link the live spec with its section anchor:

```
fix: report engine failures as errors instead of denials

OpenFGA answered an engine failure as a denial, so a client, Kong and the
logboek could not tell it from a policy denial. A failed single evaluation
now answers 500 and is logged as Error. A failed batch item stays denied and
carries the error in its context, as AuthZEN 1.0 describes:
https://openid.net/specs/authorization-api-1_0.html#section-7.2.1
```

## Code

- Names explain the code. Add a comment only for a *why* the code can't show. Don't write doc comments that repeat
  the name.
- Prefer some duplication over an early abstraction. Extract shared code once its real shape is clear, not because
  a second or third caller appears.
- Delete dead code and unused API surface. Don't keep it "just in case"; git remembers it.
- Go: one assignment per line, so no `a.x, a.y = x, y`. Functions that return several values are fine.

## Tests

- Test behaviour as integration tests against a real PostgreSQL, like `apps/pdp/server` and `apps/manager/server`.
- A test that needs a service that isn't running fails. It never skips, because a skipped run looks green.

## Branches and merge requests

- Name a branch after its Jira issue: `<issue number>-<short-slug>`, e.g. `196-context`.
- Rebase onto `main` before review, and clean up the history (fixups, order) before merging.
- The history shows the end result, not how you got there. A fix for something the branch itself introduced goes
  into the commit that introduced it, as a fixup. Only a fix for code already on `main` gets a commit of its own.
- The merge request says what changes and why, and ends with its Jira issue: `Issue: SCRUM-<issue number>`, e.g.
  `Issue: SCRUM-157`.
- Keep review comments short and concrete. Say so when an LLM wrote a comment.

## Decisions and vocabulary

- Record architecture decisions in [`docs/adr/`](docs/adr).
- Use the terms from [`CONTEXT.md`](CONTEXT.md).
