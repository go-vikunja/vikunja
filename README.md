<img src="https://vikunja.io/images/vikunja-logo.svg" alt="" style="display: block;width: 50%;margin: 0 auto;" width="50%"/>

[![Build Status](https://github.com/go-vikunja/vikunja/actions/workflows/ci.yml/badge.svg)](https://github.com/go-vikunja/vikunja/actions/workflows/ci.yml)
[![License: AGPL-3.0-or-later](https://img.shields.io/badge/License-AGPL--3.0--or--later-blue.svg)](LICENSE)
[![Install](https://img.shields.io/badge/download-v2.3.0-brightgreen.svg)](https://vikunja.io/docs/installing)
[![Docker Pulls](https://img.shields.io/docker/pulls/vikunja/vikunja.svg)](https://hub.docker.com/r/vikunja/vikunja/)
[![Swagger Docs](https://img.shields.io/badge/swagger-docs-brightgreen.svg)](https://try.vikunja.io/api/v1/docs)
[![Go Report Card](https://goreportcard.com/badge/code.vikunja.io/api)](https://goreportcard.com/report/code.vikunja.io/api)

# Vikunja

> The Todo-app to organize your life.

If Vikunja is useful to you, please consider [buying me a coffee](https://www.buymeacoffee.com/kolaente), [sponsoring me on GitHub](https://github.com/sponsors/kolaente) or buying [a sticker pack](https://vikunja.io/stickers).
I'm also offering [a hosted version of Vikunja](https://vikunja.cloud/) if you want a hassle-free solution for yourself or your team.

## Table of contents

- [Security Reports](#security-reports)
- [Features](#features)
- [Docker Workflows](#docker-workflows)
	- [Run Only (No Code Changes)](#run-only-no-code-changes)
	- [Development (Hot Reload)](#development-hot-reload)
	- [Development (Local Frontend + Docker Backend)](#development-local-frontend--docker-backend)
	- [Development (Fully Local — No Docker)](#development-fully-local--no-docker)
- [Docs](#docs)
	- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
	- [Unsplash Images](#unsplash-images)

## Security Reports

If you find any security-related issues you don't want to disclose publicly, please use [the contact information on our website](https://vikunja.io/contact/#security).

## Features

See [the features page](https://vikunja.io/features/) on our website for a more exhaustive list or 
try it on [try.vikunja.io](https://try.vikunja.io)!

## Docker Workflows

Use one of the two workflows below depending on your goal.

### Run Only (No Code Changes)

Use this when you only want to run Vikunja and stop it later.

1. Build the image:

```bash
docker build -t vikunja .
```

2. Start Vikunja:

```bash
docker run --name vikunja-run \
	-p 3456:3456 \
	-e VIKUNJA_SERVICE_PUBLICURL=http://localhost:3456/ \
	-v ~/vikunja-files:/app/vikunja/files \
	-v ~/vikunja-db:/db \
	vikunja
```

3. Open the app:

- http://localhost:3456

4. Stop and remove the container:

```bash
docker stop vikunja-run
docker rm vikunja-run
```

5. Check logs (while running):

```bash
docker logs -f vikunja-run
```

Notes:

- `~/vikunja-files` stores uploaded files.
- `~/vikunja-db` stores the SQLite database.
- If port `3456` is already in use, stop the existing container or map to a different host port (for example `-p 3457:3456`).

### Development (Hot Reload)

Use this when you plan to change code.

This repository includes a dev compose setup in `docker-compose.dev.yml`:

- Frontend runs Vite with HMR (auto-refresh in browser).
- Backend runs with Air (auto rebuild/restart on Go file changes).
- Source code is mounted into containers so edits are reflected immediately.

1. Start development stack:

```bash
docker compose -f docker-compose.dev.yml up -d
```

2. Open the app:

- Frontend (dev UI): http://localhost:4173
- Backend/API URL (for install/setup checks): http://localhost:3456

Important:

- If the UI asks for your Vikunja URL during setup, use `http://localhost:3456`.
- `http://localhost:4173` is only the Vite dev frontend.

3. View logs:

```bash
docker compose -f docker-compose.dev.yml logs -f api frontend
```

4. Stop development stack:

```bash
docker compose -f docker-compose.dev.yml down
```

5. Rebuild/restart if needed:

```bash
docker compose -f docker-compose.dev.yml restart
```

What refreshes automatically:

- Frontend changes under `frontend/`: Vite HMR updates instantly.
- Backend Go changes: Air rebuilds and restarts the API service.
- For config or dependency changes, restart the affected service.

### Development (Local Frontend + Docker Backend)

Use this when you want the fastest frontend hot-reload (native Vite HMR) while keeping the backend isolated in Docker.

1. Start only the backend:

```bash
docker compose -f docker-compose.dev.yml up -d api
```

2. Install frontend dependencies (first time only):

```bash
cd frontend && pnpm install
```

3. Start the local Vite dev server:

```bash
cd frontend && pnpm dev
```

4. Open the app:

- http://localhost:4173

How it works:

- `frontend/.env.local` sets `DEV_PROXY=http://localhost:3456`, which tells Vite to proxy all `/api` requests to the Docker backend automatically.
- No CORS issues — the browser only ever talks to `localhost:4173`.
- Backend Go changes still auto-rebuild via Air inside Docker.

5. Stop the backend when done:

```bash
docker compose -f docker-compose.dev.yml down
```

### Development (Fully Local — No Docker)

Use this when you have Go and Node.js installed and don't want Docker at all.

Requirements:

- [Go](https://go.dev/dl/) (install via `brew install go` on macOS)
- [Node.js](https://nodejs.org/) + [pnpm](https://pnpm.io/)

1. Build the frontend static files (required for the backend to compile):

```bash
cd frontend && pnpm install && pnpm build && cd ..
```

2. Start the backend:

```bash
VIKUNJA_SERVICE_PUBLICURL=http://localhost:3456/ go run . web
```

The backend starts on `http://localhost:3456` and uses SQLite by default (no separate database needed).

3. In a second terminal, start the frontend dev server:

```bash
cd frontend && pnpm dev
```

4. Open the app:

- http://localhost:4173

Note: `frontend/.env.local` must contain `DEV_PROXY=http://localhost:3456` so Vite proxies API calls to the backend. See `frontend/.env.local.example`.

For backend hot-reload on Go file changes, install [Air](https://github.com/air-verse/air) and run `air` instead of `go run . web`:

```bash
go install github.com/air-verse/air@latest
VIKUNJA_SERVICE_PUBLICURL=http://localhost:3456/ air -c .air.toml
```

## Docs

* [Installing](https://vikunja.io/docs/installing/)
* [Build from source](https://vikunja.io/docs/build-from-sources/)
* [Development setup](https://vikunja.io/docs/development/)
* [Magefile](https://vikunja.io/docs/magefile/)
* [Testing](https://vikunja.io/docs/testing/)

All docs can be found on [the Vikunja home page](https://vikunja.io/docs/).

### Roadmap

See [the roadmap](https://my.vikunja.cloud/share/QFyzYEmEYfSyQfTOmIRSwLUpkFjboaBqQCnaPmWd/auth) (hosted on Vikunja!) for more!

## Contributing

Please check out the contribution guidelines on [the website](https://vikunja.io/docs/development/).

## License

Most of this repository is licensed under [AGPL‑3.0‑or‑later](LICENSE).
The contents of [`desktop/`](desktop/) are licensed under
[GPL‑3.0‑or‑later](desktop/LICENSE).

### Unsplash Images

Background images from Unsplash are distributed under the [Unsplash License](https://unsplash.com/license). The license requires giving credit to the photographer and Unsplash. See [Unsplash’s terms](https://unsplash.com/terms) for more information.




run locally
-------------
backend:
cd /Users/mac/vikunja/vikunja
VIKUNJA_SERVICE_PUBLICURL=http://localhost:3456/ /opt/homebrew/bin/go run . web

frontend:
cd /Users/mac/vikunja/vikunja/frontend
pnpm dev