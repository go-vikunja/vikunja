# Dev commands and configuration

- `mage -l` lists targets; `pnpm run` in `frontend/` lists scripts.
- `pnpm dev` serves the frontend on port 4173 (override with `VIKUNJA_FRONTEND_PORT` or `--port`).
- Scaffolding: `mage dev:make-migration <StructName>` (prompts if omitted), `dev:make-event`, `dev:make-listener`, `dev:make-notification`. `make-listener` also adds the `events.RegisterListener(...)` call; a listener only fires if registered there.
- Invoke the `migration` skill before touching `pkg/migration/`.

## Configuration

`config.yml.sample` is generated from `config-raw.json` via `mage generate:config-yaml`. Edit the JSON, then regenerate. Environment variables override config file settings.
