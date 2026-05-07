# veans

A beans-shaped CLI for Vikunja. Drop it into a repo, run `veans init`, paste a
hook snippet into your coding agent's settings, and the agent immediately
knows to track its work in Vikunja instead of in `TodoWrite` or `.beans/`.

veans is a thin Go binary that wraps Vikunja's REST API with an opinionated
agent-friendly surface and emits a system prompt teaching agents the workflow
(claim → work → in-review → human closes). The agent prompt is re-emitted on
every `SessionStart` and `PreCompact`, so context never goes stale.

## Quick start

```sh
# 1. Build (or download) the binary
cd veans && mage build && sudo install ./veans /usr/local/bin/

# 2. In a repo with a Vikunja instance reachable
veans init --server https://vikunja.example.com

# 3. Wire it into Claude Code (.claude/settings.json):
{
  "hooks": {
    "SessionStart": [{ "hooks": [{ "type": "command", "command": "veans prime" }] }],
    "PreCompact":   [{ "hooks": [{ "type": "command", "command": "veans prime" }] }]
  }
}

# OpenCode (.opencode/plugin/veans-prime.ts):
export const VeansPrime = {
  event: ["session.start", "compact.before"],
  handler: async ({ exec }) => exec("veans prime"),
}
```

`veans prime` exits silently with status 0 when no `.veans.yml` is reachable
upward from cwd, so the hook is safe to install in a global `~/.claude/`
without breaking sessions in unrelated repos.

## What `veans init` does

1. Authenticates as you. Default is OAuth 2.0 Authorization Code + PKCE
   against Vikunja's built-in authorization server (Vikunja 2.3+ — no
   client registration needed). veans prints an authorize URL; you open
   it in your browser, sign in, and paste the resulting
   `vikunja-veans-cli://callback?code=...` URL back into the CLI. The
   browser will fail to open the custom scheme — that's expected; the
   address bar still has what we need.

   Alternative auth modes:
   - `--token <jwt-or-personal-api-token>` — paste-in, useful for SSO/OIDC
   - `--use-password` — fall back to `POST /login` (local accounts only)
   - `--username` + `--password` (non-interactive; implies `--use-password`)
2. Asks you to pick a project and a Kanban view.
3. Bootstraps the canonical buckets if missing: `Todo`, `In Progress`,
   `In Review`, `Done`, `Scrapped`.
4. Creates a `bot-<repo-name>` user (Vikunja bot user — no password, no
   email, can't log in interactively).
5. Shares the project with the bot at read+write.
6. Mints a long-lived API token for the bot via `PUT /tokens` with
   `owner_id`, scoped to the discovered route groups (tasks, comments,
   labels, relations, assignees, etc.) the server actually exposes.
7. Stores the token in your OS keychain (or
   `~/.config/veans/credentials.yml` if no keychain is available).
8. Writes `.veans.yml` to the repo root.

The token stored is the bot's, not yours. The human's transient session is
discarded as soon as init finishes — rotate or revoke the bot independently
without affecting your own session.

## Commands

```
veans init                     OAuth/login → create bot → mint token → write .veans.yml
veans prime                    emit system prompt for agents (silent if no .veans.yml)
veans list                     filtered list (--ready, --mine, --branch, --filter, --status, --json)
veans show <id>                view a task (--json for raw object)
veans create "title"           --description, --label, --status, --priority, --parent, --blocked-by
veans update <id>              --status, --title, --priority, --label-add/remove,
                               --description, --description-replace-old/new, --description-append,
                               --comment, --reason, --if-unchanged-since
veans claim <id>               assign the bot, move to In Progress, tag with current branch label
veans api METHOD PATH          raw REST passthrough — escape hatch for endpoints not wrapped here
veans login                    re-mint the bot's token (rotation)
veans version
```

Task IDs accept `PROJ-NN` (when the project has an identifier), `#NN`
(when it doesn't), or a bare integer.

## `.veans.yml`

Committed to the repo root. The numeric IDs are the source of truth; cached
identifiers and bot username are for human-readable output.

```yaml
server: https://vikunja.example.com
project_id: 42
project_identifier: PROJ        # may be "" — task IDs render as #NN then
view_id: 7
buckets:
  todo: 11
  in_progress: 12
  in_review: 13
  done: 14
  scrapped: 15
bot:
  username: bot-myrepo
  user_id: 99
```

## Credentials

Resolved in order on every command:

1. **OS keychain** (macOS Keychain, Windows Credential Manager,
   libsecret/gnome-keyring on Linux), via `github.com/zalando/go-keyring`.
2. **`VEANS_TOKEN`** env var (read-only). Optionally pin to a server with
   `VEANS_SERVER`. Intended for CI / containers.
3. **`~/.config/veans/credentials.yml`** (mode 0600) — automatic fallback
   when the keychain is unavailable. Honors `XDG_CONFIG_HOME`.

## Mage targets

```
mage build              # go build -o ./veans ./cmd/veans
mage test               # unit tests across the module
mage test:filter EXPR   # go test -run EXPR ./...
mage test:e2e           # e2e suite (needs VEANS_E2E_API_URL)
mage lint / lint:fix    # golangci-lint
mage fmt                # go fmt ./...
mage clean              # remove built binary
```

## End-to-end tests

The suite in `e2e/` assumes a running Vikunja API. Locally, point it at any
dev instance:

```sh
export VEANS_E2E_API_URL=http://localhost:3456
export VEANS_E2E_ADMIN_USER=user1
export VEANS_E2E_ADMIN_PASS=12345678   # canonical fixture password
mage test:e2e
```

CI spins Vikunja up the same way the frontend Playwright suite does — see
`.github/workflows/veans-e2e.yml`. The workflow builds the parent API
binary, starts it with `VIKUNJA_DATABASE_TYPE=sqlite`,
`VIKUNJA_DATABASE_PATH=memory`, fixtures from `pkg/db/fixtures/`, and runs
`mage test:e2e` from this directory.

E2E tests never touch the developer's keychain — they override `HOME` and
`XDG_CONFIG_HOME` per test, which forces the credential store to fall
through to its file backend.

## Status model

| Status        | Bucket name    | Done flag | Who moves there?                         |
| ------------- | -------------- | --------- | ---------------------------------------- |
| `todo`        | Todo           | false     | created here by default                  |
| `in-progress` | In Progress    | false     | `veans claim` / `update -s in-progress`  |
| `in-review`   | In Review      | false     | the agent, when work is finished         |
| `completed`   | Done           | true      | humans / merge hook only                 |
| `scrapped`    | Scrapped       | true      | the agent, with `--reason`               |

The agent never moves tasks to `completed` itself — it parks them in
`In Review` and a human (or the future merge hook) closes them once the
PR lands.

## Out of scope (for now)

- OAuth 2.0 device flow (RFC 8628) — would let SSH'd / headless setups
  authenticate without a browser-on-the-same-machine; not implemented
  upstream yet.
- Project-scoped API tokens — Vikunja doesn't ship them yet. The
  credential schema's `scope` field is forward-compatible for when it does.
- Auto-installing hook snippets. We print them; you paste them.
- Merge-hook GitHub Action that auto-closes tasks on PR merge — separate
  repo, future work.
