# AppPort bridge for Vikunja (opt-in)

A small, separate Node.js process that lets an [AppPort](https://www.npmjs.com/package/@appport/sdk) consumer discover
and call Vikunja capabilities. It does not change Vikunja: no Go code, routes, models, migrations, swagger or
configuration are touched, and nothing runs unless you start it.

| Capability | Effect | Vikunja endpoint used | Vikunja check that decides |
|---|---|---|---|
| `vikunja.projects.list@1` | read | `GET /api/v2/projects` (all pages) | `Project.CanRead`, token scope `projects.read_all` |
| `vikunja.tasks.get@1` | read | `GET /api/v2/tasks/{id}` | `Task.CanRead`, token scope `tasks.read_one` |
| `vikunja.tasks.setdone@1` | **mutation** | `PATCH /api/v2/tasks/{id}` `{"done": …}` | `Task.CanUpdate` (project write), token scope `tasks.update` |

Negative project ids in `projects.list` are Vikunja's own virtual projects (saved filters); they are passed through and
flagged `virtual: true`.

## How authority stays with Vikunja

* **Trusted-boundary mode** (`@appport/server` README): every call is forwarded to Vikunja's own `/api/v2` API with
  the **caller's own Vikunja API token** (`Authorization: Bearer tk_…`). Vikunja's route scopes, permission methods,
  update logic (`done_at`, repeating tasks) and storage decide everything.
* The bridge has **no credential of its own**, **stores nothing**, and never logs tokens or task content.
* AppPort only checks that a credential was presented (anonymous calls get `UNAUTHORIZED` without reaching Vikunja).
  Vikunja's answers map to AppPort errors: 401 → `UNAUTHORIZED` (including a token missing the needed scope),
  403 → `FORBIDDEN`, 404 → `NOT_FOUND`, 400/422 → `INVALID_INPUT`.
* The mutation uses `PATCH` on `/api/v2`, which changes only `done`. The frozen v1 `POST /tasks/{id}` is a full
  replace and would clear fields that are not sent.
* `setdone` declares, in its published contract, how to verify it (`vikunja.tasks.get` must show the requested `done`),
  that it is not idempotent at the AppPort level, and that an `INTERNAL_ERROR` means "read the task before trying again".
* Discovery is opt-in (`APPPORT_BRIDGE_PUBLIC=1`) and is never authorization.

## Run

```sh
cd appport-bridge && npm ci
APPPORT_BRIDGE_UPSTREAM=http://127.0.0.1:3456 APPPORT_BRIDGE_PORT=4420 APPPORT_BRIDGE_PUBLIC=1 npm start
```

```js
import { AppPort } from "@appport/sdk";
const app = await AppPort.discover({ origin: "http://127.0.0.1:4420", headers: { authorization: "Bearer tk_..." } });
const task = await app.call("vikunja.tasks.get", { id: 42 }, { version: 1 });
```

The bridge listens on `127.0.0.1` unless `APPPORT_BRIDGE_HOST` says otherwise. Put TLS in front of it before exposing it
beyond localhost: callers send their Vikunja tokens to it.

## Test

`npm test` starts a **real Vikunja server** (SQLite file in a temp dir; `VIKUNJA_BIN=/path/to/vikunja` to reuse a build,
otherwise `go build -tags osusergo .`), creates synthetic users, projects, tasks and scoped API tokens through Vikunja's
own API, starts the bridge, and checks everything as an independent consumer using only published `@appport` packages:

* discovery is opt-in and `reference-verified`; it establishes no trust, authorization or approval
* every successful result is compared with **Vikunja's own `projects` / `tasks` tables, read directly** (read-only)
* carol cannot read or change bob's task (`FORBIDDEN`; the database row is unchanged); a read-only token cannot change a
  task (Vikunja answers 401 → `UNAUTHORIZED`; unchanged); no token → `UNAUTHORIZED`
* a missing task is `NOT_FOUND`; malformed input is `INVALID_INPUT`
* `setdone` is confirmed by the contract's declared read-back (published `@appport/integration`) **and** the database,
  with other task fields untouched
* a lost `setdone` response is *indeterminate*, is sent once and never retried; only the database shows it ran
* all disposable data is removed

Package versions are pinned exactly (`@appport/sdk` 1.2.2; test-only `@appport/protocol` 1.1.0 and
`@appport/integration` 1.1.0) and installed from the npm registry via `package-lock.json`. Requires Node 22.13+.
