# Dev commands and configuration

- `mage -l` lists targets; `pnpm run` in `frontend/` lists scripts.
- `pnpm dev` serves the frontend on port 4173 (override with `VIKUNJA_FRONTEND_PORT` or `--port`).
- Scaffolding: `mage dev:make-migration <StructName>` (prompts if omitted), `dev:make-event`, `dev:make-listener`, `dev:make-notification`. `make-listener` also adds the `events.RegisterListener(...)` call; a listener only fires if registered there.
- Invoke the `migration` skill before touching `pkg/migration/`.

## Configuration

`config.yml.sample` is generated from `config-raw.json` via `mage generate:config-yaml`. Edit the JSON, then regenerate. Environment variables override config file settings.

## Frontend API URL

- `window.API_URL` and `localStorage.API_URL` hold the server root without `/api/vN`; `''` means the frontend's own origin. `frontend/index.html` ships `''`; `pkg/routes/static.go` replaces it with `service.publicurl`.
- Build request URLs with `getApiBaseUrl()` (appends `/api/v2`) or `getApiRootUrl()` from `frontend/src/helpers/apiUrl.ts`, never from `window.API_URL`. `normalizeApiUrl()` strips a trailing `/api/vN` from any input.
- Change it only through `checkAndSetApiUrl()`: it probes `/info`, then reconfigures the client and clears the query cache.
