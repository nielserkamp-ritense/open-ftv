# Issue tracker: Jira

Issues and PRDs for this repo live in Jira, project **FTV - ICTU**.

## Conventions

- **Create an issue**: use the `jira` CLI. Typical shape: `jira issue create -p"<PROJECT_KEY>" -t"Task" -s"<summary>" -b"<description>"`. Use a heredoc for multi-line descriptions and pass `-b -` (or the CLI's stdin equivalent) when you need to pipe in markdown. Confirm the exact project key for "FTV - ICTU" with the user the first time it's needed if it isn't already configured in `~/.config/.jira/.config.yml`.
- **Read an issue**: prefer the **Jira MCP** if available in this session. Fall back to the Jira REST API (`GET /rest/api/3/issue/<KEY>?expand=renderedFields`) when MCP isn't connected.
- **List / search issues**: use the Jira MCP, or `GET /rest/api/3/search?jql=...`. Build JQL like `project = "FTV" AND labels = "ready-for-agent"`.
- **Comment on an issue**: `jira issue comment add <KEY> "<body>"`, or `POST /rest/api/3/issue/<KEY>/comment` via the API/MCP.
- **Apply / remove labels**: `jira issue edit <KEY> --label "..."` / `--label-remove "..."`, or the API equivalent.
- **Transition state**: Jira uses workflow transitions, not just labels. Use `jira issue move <KEY> "<Transition name>"` or the API's `/transitions` endpoint. Triage states are tracked as labels (see `triage-labels.md`), not as Jira statuses, unless the user states otherwise later.
- **Close**: transition to the project's "Done" / "Closed" status; post any explanation as a comment first.

There are no project-specific conventions beyond the defaults — labels map 1:1 to the canonical triage vocabulary.

## When a skill says "publish to the issue tracker"

Create a Jira issue in project **FTV** (ICTU instance) using the `jira` CLI.

## When a skill says "fetch the relevant ticket"

Use the Jira MCP if connected; otherwise call the Jira REST API. Ask the user for the issue key if it isn't obvious from context.
