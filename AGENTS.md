# AGENTS.md

## Agent skills

### Issue tracker

Issues live in Jira project **FTV - ICTU** (ICTU). Agents create issues via the `jira` CLI and read via the Jira API / MCP. See `docs/agents/issue-tracker.md`.

### Triage labels

Canonical label vocabulary used as-is (no remapping). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.

### Working conventions

Commits, code style, tests and merge requests follow CONTRIBUTING.md. It is imported here, so every session and
subagent starts with its rules:

@CONTRIBUTING.md
