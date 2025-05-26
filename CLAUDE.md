# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

Vikunja is a todo/task management application with three main components:

- **Go Backend** (`/pkg/`): RESTful API server using Echo framework, XORM for database operations, supports SQLite/MySQL/PostgreSQL
- **Vue.js Frontend** (`/frontend/`): Vue 3 + TypeScript + Vite, Pinia state management, multiple view types (List, Kanban, Gantt, Table)
- **Desktop App** (`/desktop/`): Electron wrapper serving the frontend locally

## Development Commands

### Backend (Go)
```bash
# Build and run
mage build                    # Build binary
./vikunja		              # Run development server

# Testing
mage test:unit               # Unit tests  
mage test:integration        # Integration tests
mage test:all			     # Run all tests

# Code quality
mage lint                    # Run golangci-lint
mage lint:fix                # Auto-fix linting issues

# Development tools
mage dev:make-migration      # Create new database migration
mage generate:swagger-docs   # Generate API documentation
```

### Frontend (Vue.js)
```bash
cd frontend/
pnpm install                 # Install dependencies
pnpm dev                     # Development server
pnpm build                   # Production build
pnpm lint                    # ESLint
pnpm lint:fix                # Auto-fix linting issues
pnpm test:unit               # Unit tests (Vitest)
pnpm test:e2e                # E2E tests (Cypress)
pnpm typecheck               # TypeScript checking
```

### Desktop App
```bash
cd desktop/
pnpm start                   # Run Electron app
pnpm dist                    # Build distributables
```

## Key Development Patterns

### Backend
- Use Mage as the build tool (like Make but in Go)
- Database migrations in `/pkg/migration/` with timestamp-based naming
- Services follow repository pattern in `/pkg/models/`
- API routes in `/pkg/routes/api/`
- Event-driven architecture with listeners in various `/pkg/*/listeners.go` files

### Frontend
- Service/Model architecture pattern (documented in `/frontend/docs/models-services.md`)
- Services in `/frontend/src/services/` handle API communication
- Models in `/frontend/src/models/` define data structures
- Stores in `/frontend/src/stores/` use Pinia for state management
- Components organized by feature in `/frontend/src/components/`

## Database & Migrations

- Use `mage dev:make-migration` to create new database migrations
- Migrations are automatically applied on startup
- Supports SQLite (default), MySQL, and PostgreSQL
- Test fixtures in `/pkg/db/fixtures/` for integration tests

## Testing Strategy

### Backend
- Unit tests throughout `/pkg/` modules using Go's testing package
- Integration tests in `/pkg/integrations/` test full API endpoints
- Run tests with `mage test:unit` and `mage test:integration`

### Frontend  
- Unit tests with Vitest (`pnpm test:unit`)
- E2E tests with Cypress (`pnpm test:e2e`)
- Test factories in `/frontend/cypress/factories/` for generating test data

## API Documentation

- Swagger documentation available at `/api/v1/docs` when running the server
- Generate docs with `mage generate:swagger-docs` after updating API annotations
- API follows RESTful conventions with standard HTTP methods

## Frontend Development Setup

Set the `DEV_PROXY` environment variable to point to a backend instance:
```bash
export DEV_PROXY=https://try.vikunja.io
cd frontend/ && pnpm dev
```

### Linting

When checking for lint issues, always run `pnpm lint:fix`, then fix only those issues which eslint couldn't fix on its own.

### SCSS Variables

You do not need to fix usage of $radius scss variables. They are globally available, even if the linter says otherwise.

## Common File Locations

- **Config**: `/config.yml.sample` (generate with `mage generate:config-yaml`)
- **Database models**: `/pkg/models/`
- **API routes**: `/pkg/routes/api/`
- **Frontend components**: `/frontend/src/components/`
- **Vue stores**: `/frontend/src/stores/`
- **Translations**: `/pkg/i18n/lang/` and `/frontend/src/i18n/`
