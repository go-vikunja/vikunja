# AGENT Instructions

## Project Overview

Vikunja is a comprehensive todo and task management application with a Vue.js frontend and Go backend. It supports multiple project views (List, Kanban, Gantt, Table), team collaboration, file attachments, and extensive integrations.

The project consists of:
- `pkg/` – Go code for the API service
- `frontend/` – Vue.js based web client
- `magefile.go` – Mage build script providing tasks for development and release
- `desktop/` – Electron wrapper application
- `docs/` – Documentation website

## Development Commands

### Backend (Go)
- **Build**: `mage build` - Builds the Go binary
- **Test Features**: `mage test:feature` - Runs feature tests
- **Test Web**: `mage test:web` - Runs web tests
- You can run specific tests with `mage test:filter <filter>` where `<filter>` is a go test filter string.
- **Lint**: `mage lint` - Runs golangci-lint
- **Lint Fix**: `mage lint:fix` - Runs golangci-lint with auto-fix
- **Generate Swagger Docs**: `mage generate:swagger-docs` - Updates API documentation (Generally you won't need to run this unless the user tells you to. It is updated automatically in the CI workflow)
- **Check Swagger**: `mage check:got-swag` - Verifies swagger docs are up to date
- **Generate Config**: `mage generate:config-yaml` - Generate sample config from `config-raw.json`
- **Clean**: `mage build:clean` - Cleans build artifacts
- **Format**: `mage fmt` - Format Go code before committing

-Development helpers under the `dev` namespace:
- **Migration**: `mage dev:make-migration <StructName>` - Creates new database migration. If you omit `<StructName>`, the command will prompt for it.
- **Event**: `mage dev:make-event` - Create an event type
- **Listener**: `mage dev:make-listener` - Create an event listener  
- **Notification**: `mage dev:make-notification` - Create a notification skeleton

### Frontend (Vue.js)
Navigate to `frontend/` directory:
- **Dev Server**: `pnpm dev` - Starts development server
- **Build**: `pnpm build` - Production build
- **Build Dev**: `pnpm build:dev` - Development build  
- **Lint**: `pnpm lint` - ESLint check
- **Lint Fix**: `pnpm lint:fix` - ESLint with auto-fix
- **Lint Styles**: `pnpm lint:styles` - Stylelint check for CSS/SCSS
- **Lint Styles Fix**: `pnpm lint:styles:fix` - Stylelint with auto-fix
- **Type Check**: `pnpm typecheck` - Vue TypeScript checking
- **Test Unit**: `pnpm test:unit` - Vitest unit tests
- **Test E2E**: `pnpm test:e2e` - Cypress end-to-end tests
- **Test E2E Dev**: `pnpm test:e2e-dev` - Interactive Cypress testing

### Pre-commit Checks
Always run both lint before committing:
```bash
# Backend
mage lint:fix

# Frontend  
cd frontend && pnpm lint:fix && pnpm lint:styles:fix
```

Fix any errors the lint commands report, then try comitting again.

You only need to run the lint for the backend when changing backend code, and the lint for the frontend only when changing frontend code. Similarly, only run style linting when modifying CSS/SCSS files or Vue component styles.

## Architecture Overview

### Backend Architecture (Go)
The Go backend follows a layered architecture with clear separation of concerns:

**Core Layers:**
- **Models** (`pkg/models/`) - Domain entities with business logic and CRUD operations
- **Services** (`pkg/services/`) - Business logic layer handling complex operations
- **Routes** (`pkg/routes/`) - HTTP API endpoints and routing configuration
- **Web** (`pkg/web/`) - Generic CRUD handlers and web framework abstractions

**Key Patterns:**
- **Generic CRUD**: Models implement `CRUDable` interface for standardized database operations
- **Permissions System**: Three-tier permissions (Read/Write/Admin) enforced across all operations
- **Event-Driven**: Event system for notifications, webhooks, and cross-cutting concerns
- **Modular Design**: Pluggable authentication, avatar providers, migration tools

**Database:**
- XORM ORM with support for MySQL, PostgreSQL, SQLite
- Migration system in `pkg/migration/` with timestamped files
- Database sessions with automatic transaction handling

**Authentication:**
- Multi-provider: Local, LDAP, OpenID Connect
- JWT tokens for API access
- API tokens with scoped permissions
- TOTP/2FA support

### Frontend Architecture (Vue.js)
Modern Vue 3 composition API application with TypeScript:

**State Management:**
- **Pinia** stores in `src/stores/` for global state
- Composables in `src/composables/` for reusable logic
- Component-level state with Vue 3 Composition API

**Key Directories:**
- `src/components/` - Reusable Vue components organized by feature
- `src/views/` - Page-level components and routing
- `src/stores/` - Pinia state management
- `src/services/` - API service layer matching backend models
- `src/models/` - TypeScript interfaces matching backend models
- `src/helpers/` - Utility functions and business logic

**UI Framework:**
- Bulma CSS framework with CSS variables for theming
- FontAwesome icons with tree-shaking
- TipTap rich text editor for task descriptions
- Custom component library in `src/components/base/`

## Development Workflows

### Adding New Features

**Backend Changes:**
1. Create/modify models in `pkg/models/` with proper CRUD and Permissions interfaces as required
2. Add database migration if needed: `mage dev:make-migration <StructName>`
3. Create/update services in `pkg/services/` for complex business logic
4. Add API routes in `pkg/routes/api/v1/` following existing patterns
5. Update Swagger annotations

**Frontend Changes:**
1. Create TypeScript interfaces in `src/modelTypes/` matching backend models
2. Add/update services in `src/services/` for API communication
3. Create components in appropriate `src/components/` subdirectories
4. Add views/pages in `src/views/` with proper routing
5. Update Pinia stores if global state changes are needed

### Database Changes
1. Run `mage dev:make-migration <StructName>`
2. Edit the generated migration file in `pkg/migration/`
3. Update corresponding model in `pkg/models/`
4. Update TypeScript interfaces in frontend `src/modelTypes/`

### API Development
- All API endpoints follow RESTful conventions under `/api/v1/`
- Use generic web handlers in `pkg/web/handler/` for standard CRUD operations
- Implement proper permissions checking using the Permissions interface
- Add Swagger annotations for automatic documentation generation

### Testing
- Backend: Feature tests alongside source files, web tests in `pkg/webtests/`
- Frontend: Unit tests with Vitest, E2E tests with Cypress
- Always test both positive and negative authorization scenarios
- Use test fixtures in `pkg/db/fixtures/` for consistent test data

## Swagger API Documentation

Never touch the generated swagger api documentation under `pkg/swagger/`. These are automatically generated by CI after committing.

## Commit Messages

Use the **Conventional Commits** style when committing changes (for example, `feat: add foo` or `fix: correct bar`). This repository uses these messages to generate changelogs.

## Frontend Development Guidelines

The web client lives in `frontend/` and uses Vue 3 + TypeScript. ESLint rules enforce: single quotes, trailing commas, no semicolons, tab indent, Vue <script lang="ts">, PascalCase component names, camelCase events. See `frontend/eslint.config.js` and `frontend/.editorconfig` and obey formatting rules outlined there.

## Translations

When adding or changing functionality which touches user-facing messages, these need to be translated.

In the frontend, all translation strings live in `frontend/src/i18n/lang`. For the api (which mainly affects the localization of notifications), the strings live in `pkg/i18n/lang`.

You only need to adjust the `en.json` file with the source string. The actual translation happens elsewhere.
After adjusting the source string, you need to call the respective translation library with the key. Both are similar, check the existing code to figure it out.

## Key Files and Conventions

**Configuration:**
- `config.yml.sample` - Example configuration (generated from `config-raw.json`)
- Environment variables override config file settings
- Use `pkg/config/` for configuration management

**Code Style:**
- Go: golangci-lint per `.golangci.yml`; use goimports; wrap errors with `fmt.Errorf("...: %w", err)`; enforce permissions checks in models; never log secrets; do not edit generated `pkg/swagger/*`
- Vue: ESLint + TS; single quotes, trailing commas, no semicolons, tab indent; script setup + lang ts; keep services/models in sync with backend
- Follow existing patterns for consistency

**Naming Conventions:**
- Go: Standard Go conventions (PascalCase for exports, camelCase for private)
- Vue: PascalCase for components, camelCase for composables
- API endpoints: kebab-case in URLs, camelCase in JSON

**Permissions and Permissions:**
- Always implement Permissions interface for new models
- Use `CanRead`, `CanWrite`, `CanCreate`, `CanDelete` methods
- Permissions are enforced at the model level, not just routes

## Common Gotchas

- Database migrations are irreversible in production - test thoroughly
- Frontend services must match backend model structure exactly
- Permissions checking is mandatory for all CRUD operations
- Event listeners in `pkg/*/listeners.go` must be registered properly
- CORS settings in backend must allow frontend domain
- API tokens have different scopes - check permissions carefully

