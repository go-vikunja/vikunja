---
name: veans
description: Use when tracking tasks in Vikunja via the veans CLI in a repo that has a .veans.yml — claiming/creating/updating tasks, the claim→work→in-review workflow, writing HTML task descriptions and comments, and reading veans JSON output. Replaces TodoWrite for task tracking in these repos.
user-invocable: true
---

# Tracking work in Vikunja via veans

`veans` is a CLI that wraps Vikunja's REST API with an agent-friendly surface (the tool lives in `veans/` in this monorepo). In any repo that has a `.veans.yml` at its root, **track your work in Vikunja with `veans` instead of `TodoWrite`** — tasks then stay visible across sessions and to the humans collaborating on the project.

If there is no `.veans.yml` reachable upward from the working directory, this skill does not apply — fall back to `TodoWrite`.

## First: prime yourself with the repo's config

Run this once at the start to get the repo-specific project, view, bucket IDs, and your bot identity:

```sh
veans prime
```

`veans prime` emits the canonical agent prompt with this repo's concrete values filled in (project ID, Kanban view ID, bucket IDs, bot username, label namespace). It exits silently with status 0 when no `.veans.yml` is found, so it's always safe to run. Treat its output as authoritative — the rest of this skill is the version-independent summary.

## Workflow: claim → work → in-review → human closes

**Before you start work:**
- Find something ready: `veans list --ready` (Todo + not blocked).
- If a task exists, claim it: `veans claim <id>` — assigns you the bot, moves it to In Progress, and tags it with the current git branch.
- Otherwise create and start in one step: `veans create "<short title>" -s in-progress -d "<p>HTML description</p>"`.

**While you work:**
- Keep the description in sync. Append a step list, or check items off with surgical replaces:
  ```sh
  veans update <id> --description-append '<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>step 1</p></li></ul>'
  veans update <id> --description-replace-old 'data-checked="false"><p>step 1</p>' --description-replace-new 'data-checked="true"><p>step 1</p>'
  ```
- Comment on significant decisions or course-changes: `veans update <id> --comment '<p>Discovered Y; pivoting to Z because …</p>'`.
- For work that could be assigned separately, create real subtasks with `--parent <id>`. For incremental checklists, use task-list items in the description instead.

**After you finish work:**
- Move to `in-review` with a summary comment. **Never move a task to `completed` yourself** — a human or the merge hook closes it once the PR lands.
  ```sh
  veans update <id> -s in-review --comment '<h3>Summary of changes</h3><ul><li>first thing</li><li>second thing</li></ul>'
  ```
- If you abandon the work, scrap it with a reason: `veans update <id> -s scrapped --reason "obsolete: <why>"`.

**Commit messages:** include the task identifier on a `Refs:` line so the merge hook can auto-close on merge:

```
fix: handle empty project identifiers

Refs: PROJ-12
```

## Status model

| Status        | Bucket      | Done | Who moves it there                       |
| ------------- | ----------- | ---- | ---------------------------------------- |
| `todo`        | Todo        | no   | created here by default                  |
| `in-progress` | In Progress | no   | `veans claim` or `update -s in-progress` |
| `in-review`   | In Review   | no   | you, when work is finished               |
| `completed`   | Done        | yes  | **humans / merge hook only**             |
| `scrapped`    | Scrapped    | yes  | you, with `--reason`                     |

## Descriptions and comments are HTML — not markdown

Vikunja renders these fields through the TipTap editor, which stores HTML. Markdown saves as literal text and looks broken in the UI. Write HTML directly. **Titles, however, are plaintext** — no tags, no markdown (they leak into list views and notifications as escaped entities).

Canonical TipTap shapes that render cleanly:

```html
<h2>Summary</h2>
<p>Short paragraph.</p>

<h3>Steps</h3>
<ul data-type="taskList">
  <li data-type="taskItem" data-checked="false"><p>find the bug</p></li>
  <li data-type="taskItem" data-checked="true"><p>write the test</p></li>
</ul>

<ul><li>plain bullet</li></ul>
<p>Inline <code>code</code>, <strong>bold</strong>, <a href="https://example.com">link</a>.</p>
<pre><code class="language-go">if err != nil { return err }</code></pre>
<blockquote><p>A quote.</p></blockquote>
```

Rules that bite:
- Interactive checkboxes **require** `<ul data-type="taskList">` + `<li data-type="taskItem" data-checked="true|false">`. Plain `<ul><li>` renders as static bullets.
- Inner text of a task item must sit inside `<p>` — the editor expects block content in the `<li>`.
- Don't add `data-task-id` attributes; the editor auto-fills them on first save.
- Escape literal `<`, `>`, `&` as `&lt;`, `&gt;`, `&amp;` (including inside `<pre><code>`).
- `--description-replace-old` matches raw HTML byte-for-byte. Make the `old` string unique by including surrounding tags, or it errors (same semantics as the Edit tool).

## Output: always JSON

Every `list`, `show`, `create`, `update`, `claim`, and `api` call emits JSON on stdout — no `--json` flag, no human-formatted variant. `list` returns an array; the others return a single task. Parse it.

Errors land on **stderr** as `{"code":"...","error":"..."}` with a non-zero exit. Branch on the stable `code`: `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`, `AUTH_ERROR`, `RATE_LIMITED`, `NOT_CONFIGURED`, `BOT_USERS_UNAVAILABLE`, `UNKNOWN`.

Useful task fields: `id` (numeric, internal — pass to `api`), `index` (per-project number behind `PROJ-NN`), `title`, `description` (HTML), `done`, `priority`, `buckets[]` (current bucket per view — match `project_view_id` to yours from `.veans.yml`), `assignees[]`, `labels[]`.

Task IDs accept `PROJ-NN`, `#NN` (when the project has no identifier), or a bare integer.

## Common commands

```sh
veans list                          # all tasks, tree view
veans list --ready                  # Todo + not blocked
veans list --mine                   # assigned to you
veans list --branch                 # tagged with the current git branch
veans list --filter "priority > 3"  # raw Vikunja filter expression
veans show <id>                     # full task detail

veans create "title" -s in-progress -d "<p>HTML body</p>"
veans create "title" --label bug --priority 4 --parent <id>
veans create "title" --blocked-by <id>

veans update <id> -s in-review --comment '<p>Summary…</p>'
veans update <id> --label-add bug --label-remove flaky
veans update <id> --description-append '<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>new step</p></li></ul>'
veans update <id> -s scrapped --reason "obsolete: replaced by PROJ-9"
veans update <id> --if-unchanged-since <ts>   # optimistic concurrency; CONFLICT if changed

veans claim <id>                    # assign yourself + In Progress + branch label
veans api GET /tasks/123            # raw REST escape hatch for endpoints not wrapped here
```

Labels live under the `veans:` namespace (auto-prepended, so `--label bug` becomes `veans:bug`); branch labels are `veans:branch:<branch>`, which `veans claim` adds automatically.

## Setup is a human task

`veans init` (bootstrap a repo: pick project + view, create the bot, mint a token, write `.veans.yml`, install hooks) and `veans login` (rotate the bot token) are run by a human, not an agent. If `veans` reports `NOT_CONFIGURED`, ask the human to run `veans init` rather than attempting it yourself.
