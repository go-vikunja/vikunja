# AGENT Instructions for veans

Things to know before touching this submodule that aren't obvious from
reading the code. The parent repo's `CLAUDE.md` covers the rest of Vikunja;
this file is veans-specific.

## Module layout

- `veans/` is its own Go module (`code.vikunja.io/veans`), separate from
  the parent. Don't try to import `code.vikunja.io/api/...` — that pulls
  XORM into the CLI binary. Wire types live in `internal/client/types.go`
  as plain JSON-tagged structs that mirror the parent models.
- License headers are enforced by `goheader` in `veans/.golangci.yml`.
  Every new `.go` file needs the AGPLv3 banner from
  `veans/code-header-template.txt` (a copy of the parent's, kept local
  so the linter resolves the path relative to this module).

## Building and testing

- `mage build` → `./veans` binary. The `Aliases` map in `magefile.go`
  routes bare names like `mage test` to `Test.All` — without aliases,
  mage rejects namespace invocations ("Unknown target specified").
- Unit tests: `mage test` or `go test ./...`.
- E2e tests: assume an externally-running Vikunja at `VEANS_E2E_API_URL`
  and admin creds in env (`VEANS_E2E_ADMIN_TOKEN`, or
  `VEANS_E2E_ADMIN_USER` + `VEANS_E2E_ADMIN_PASS`). The package
  self-skips when `VEANS_E2E_API_URL` is empty, so plain `go test` is
  safe locally.
- Local e2e loop: from the parent repo root, build the API
  (`mage build:build`), run it with sqlite-memory + a known JWT secret,
  register an admin user via `POST /register`, then
  `go test ./e2e/...` from `veans/` with the env vars above.
- CI: the `test-veans-e2e` job in `.github/workflows/test.yml` consumes
  the existing `vikunja_bin` artifact from `api-build`; don't recompile
  the API in a parallel workflow. The `veans-test` job runs unit tests
  independently and gives fast feedback.

## Vikunja wire-format gotchas

Most failures surface when crossing the JSON boundary. The list below is
what's bitten me; if a new endpoint behaves oddly, suspect one of these:

- **`ProjectView.view_kind` and `bucket_configuration_mode` are
  strings**, not ints. The parent enums (`ProjectViewKind`,
  `BucketConfigurationModeKind`) have custom `MarshalJSON` that emits
  `"kanban"` / `"manual"` etc. Use the string constants in
  `internal/client/types.go`.
- **`Task.BucketID` is always 0** in `GET /tasks/:id`. The model has
  `xorm:"-"` on it — the actual bucket lives in a separate
  `task_buckets` table. Fetch with `?expand=buckets` and use
  `task.CurrentBucketID(viewID)` to read it.
- **`POST /tasks/{id}` does NOT move tasks between buckets.** The
  task↔bucket relation is row-shaped; use `client.MoveTaskToBucket()`
  which hits `POST /projects/{p}/views/{v}/buckets/{b}/tasks`. The
  Update path on the server only auto-moves on `done` flips.
- **Bot user creation is `PUT /user/bots`**, not `/bots` — the routes
  are registered under the `/user` subgroup. Same prefix for
  `GET /user/bots`.
- **`APIToken.expires_at` is required.** The struct field has
  `valid:"required"` upstream; sending it omitted or zero fails
  validation. Use `client.FarFuture` (year 9999) when you mean "no
  expiry" — the frontend does the same.
- **Task descriptions and comments are HTML, not markdown.** The
  Vikunja web UI uses TipTap, which calls `getHTML()` on save. The
  stored field is therefore HTML. The agent prompt template
  (`internal/commands/prompt.tmpl`) teaches agents the canonical
  TipTap shapes — most importantly `<ul data-type="taskList">` +
  `<li data-type="taskItem" data-checked="false"><p>…</p></li>` for
  interactive checkboxes. We deliberately do **not** convert
  markdown↔HTML in the CLI; the agent writes HTML directly, which
  avoids lossy roundtrips on `--description-replace-old/new`. `veans
  show` displays the raw HTML; humans skim it fine.

## API token permissions

- Vikunja validates token `permissions` against `apiTokenRoutes`, a map
  built dynamically from registered routes. Group names are derived
  from the URL path (params stripped, joined by `_`). Examples:
  - `/projects/:project/views/:view/buckets/:bucket/tasks` →
    group `projects`, action `views_buckets_tasks`
  - `/tasks/:task/comments` → group `tasks_comments`, action `create`
- `client.PermissionsForBot()` calls `GET /routes` at runtime and
  grants only the intersection of what we want and what the server
  exposes. **Don't hard-code permission group names** — they drift
  across Vikunja versions, and discovery keeps the bot's grant valid
  across upgrades.

## Bot ownership and token minting

- Creating a bot via `PUT /user/bots` automatically sets the bot's
  `bot_owner_id` to the calling user. Only the owner can mint tokens
  for the bot via `PUT /tokens` with `owner_id=<bot_id>`. The init
  flow does these as a single human-JWT-authenticated batch.
- Bots have no password and **cannot** authenticate via `POST /login`.
  After init, `veans login` re-authenticates as the human (not the
  bot) and mints a fresh bot token.

## OAuth flow

- Vikunja's authorization server requires PKCE/S256 and accepts either
  `vikunja-…://` custom schemes or RFC 8252 loopback URIs
  (`http://127.0.0.1:NNN/`, `http://localhost:NNN/`, `http://[::1]:NNN/`).
  No client registration needed — `client_id` can be any consistent
  string (we use `veans-cli`).
- `internal/auth/oauth.go` binds a free port on 127.0.0.1, opens the
  browser, and captures the callback. The `Shutdown` defer uses
  `context.WithoutCancel(ctx)` so cancellation at the outer scope
  still drains the loopback server cleanly.
- Token exchange is **JSON only**. Form-encoded POSTs to `/oauth/token`
  fail; the standard `golang.org/x/oauth2` client speaks form encoding,
  which is why we have a hand-rolled `client.ExchangeOAuthCode`.

## Credential store

- Lookup chain: keychain → env (`VEANS_TOKEN`, optionally pinned by
  `VEANS_SERVER`) → file (`~/.config/veans/credentials.yml`, mode 0600,
  honors `XDG_CONFIG_HOME`).
- `Chain.Set` falls through to the next backend on error so a missing
  dbus on a CI runner doesn't block writes — the file backend is the
  reliable last-resort.
- E2e tests override `HOME` and `XDG_CONFIG_HOME` per test to keep the
  developer's keyring untouched. Don't bypass the credentials package
  in tests — leaks between tests will surface as the wrong bot token.

## Project identifiers and bot usernames

- Project `Identifier` is `runelength(0|10)`, can be empty. When empty,
  `Config.FormatTaskID` renders `#NN`; otherwise `PROJ-NN`. Both are
  accepted by `runtime.resolveTaskID` along with bare integers.
- Bot username must start with `bot-`; the server enforces it. Hyphens,
  digits, lowercase letters allowed; no spaces, no commas, no
  `link-share-N` pattern. `config.SuggestedBotUsername` does the
  folding for repo names.
- E2e tests deriving identifiers from a unique suffix should use the
  trailing chars of `strconv.FormatInt(time.Now().UnixNano(), 36)`.
  The leading chars barely change between consecutive runs and will
  collide if you take `[:N]`.

## Audience split

The CLI is agent-only at runtime; humans never use it for day-to-day
work (they use Vikunja's web UI). Two commands serve a human running
one-off setup:

- **`init`** — bootstrap a repo: pick project + view, create bot,
  share, mint token, write `.veans.yml`, install hooks.
- **`login`** — rotate the bot's token.

Everything else (`list`, `show`, `create`, `update`, `claim`, `api`,
`prime`, `version`) is **agent-only**:

- **Emits JSON on stdout unconditionally.** No `--json` flag, no
  human-formatted variant. `list` is a raw array; `show` / `create` /
  `update` / `claim` return a single task object.
- **Errors are JSON on stderr** with non-zero exit — same envelope
  everywhere (`{"code": "...", "error": "..."}`), regardless of which
  command ran. Stable codes in `internal/output/errors.go`:
  `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`, `AUTH_ERROR`,
  `RATE_LIMITED`, `BOT_USERS_UNAVAILABLE`, `NOT_CONFIGURED`,
  `UNKNOWN`. Don't add ad-hoc strings — wrap with `output.New` /
  `output.Wrap`.
- **No `globals.JSON`, no dual rendering paths.** If you find yourself
  reaching for "if interactive, do X" on an agent-facing command,
  stop — it's not interactive, an agent is on the other end.

## Cobra surface conventions

- `RunE` handlers that don't use `args []string` should rename it to
  `_` to satisfy revive's `unused-parameter` rule.
- The bucket-move dance (`MoveTaskToBucket`) runs **after** the field
  update on `update`, so a status transition can't clobber freshly
  attached labels. Comments for `--status scrapped` post **before**
  the bucket move so the audit trail reads in chronological order.
- Agent-facing commands return the task via `json.NewEncoder(...).Encode(task)`.
  Adding new top-level keys to `client.Task` is an implicit API
  change — bump `prime`'s "useful fields" note alongside.

## Things to *not* do

- **Don't add an `os/exec.Command`** without ctx — `noctx` is enabled.
  Use `exec.CommandContext(ctx, …)` and thread the context through.
- **Don't commit the built binary.** `veans/.gitignore` covers
  `./veans` and `./veans.exe`.
- **Don't write to stdout from `prime` when no `.veans.yml` is found.**
  The hook contract is silent + exit 0 so the snippet is safe to install
  globally in `~/.claude/settings.json`.
- **Don't change canonical bucket titles** without updating
  `internal/status/CanonicalBucketTitles`, the prompt template, and
  the e2e assertions in lockstep — agents and humans both treat them
  as fixed strings.
