---
name: start-dev-server
description: >-
  Starts the local Vikunja development stack (API + frontend Vite server).
  Use when the user says "start dev server", "start the dev server",
  "run local dev", or invokes this skill explicitly.
---

# Start local Vikunja dev server

Bring up the API and frontend for local development. Do not ask for confirmation unless a required prerequisite is missing.

## Expected endpoints

| Service  | URL                        | How it runs                          |
|----------|----------------------------|--------------------------------------|
| API      | `http://127.0.0.1:3456`    | `./vikunja` from repo root           |
| Frontend | `http://127.0.0.1:4173/`   | `pnpm dev` in `frontend/`            |

Frontend proxies API calls via `frontend/.env.local`:

```
DEV_PROXY=http://127.0.0.1:3456
```

## Workflow

1. **Check what is already running**
   - Listeners: API `:3456`, frontend `:4173`
   - Reuse anything already healthy; only start missing pieces

2. **Ensure frontend proxy config**
   - If `frontend/.env.local` is missing `DEV_PROXY`, create/update it to `DEV_PROXY=http://127.0.0.1:3456`
   - Do not commit `.env.local`

3. **Start API if needed**
   - Prerequisites: `config.yml` at repo root; `vikunja` binary present (run `mage build` if missing)
   - From repo root, background:
     ```bash
     ./vikunja 2>&1 | tee /tmp/vikunja-api.log
     ```
   - Wait until logs show the server is listening (or `:3456` is open)
   - Use full permissions (`all`) so the process can bind and write its DB/files

4. **Start frontend if needed**
   - Prerequisites: `frontend/node_modules` (run `pnpm install` in `frontend/` if missing)
   - From `frontend/`, background:
     ```bash
     pnpm dev 2>&1 | tee /tmp/vikunja-frontend.log
     ```
   - Wait for Vite `ready` / `Local: http://127.0.0.1:4173/`
   - Use full permissions (`all`); sandbox often breaks `pnpm`/`vite`

5. **Report to the user**
   - Frontend URL: `http://127.0.0.1:4173/`
   - API URL: `http://127.0.0.1:3456`
   - Note if a service was already running vs newly started
   - On failure, show the relevant log tail from `/tmp/vikunja-api.log` or `/tmp/vikunja-frontend.log`

## Do not

- Run E2E (`mage test:e2e`) or production builds for this task
- Kill healthy existing servers unless the user asks to restart
- Commit `.env.local`, `vikunja.db*`, or binary artifacts
