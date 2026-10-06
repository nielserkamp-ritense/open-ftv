# Contributing

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
