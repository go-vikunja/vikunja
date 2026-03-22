# Project-Scoped API Tokens

## Overview

Add the ability to scope API tokens to a specific project (and optionally its sub-projects). When a token is project-scoped, it can only access resources within that project's scope — enforced through the existing permission system by carrying scope information on the User object.

## Design

### Core Idea

Instead of adding a separate middleware layer for scope enforcement, we **extend the User struct** with transient scope fields. The token middleware sets these fields when a project-scoped token is used. The existing permission methods (`CanRead`, `CanWrite`, etc.) then type-assert to a `ProjectScoped` interface and reject access to out-of-scope resources.

This means scope enforcement happens in exactly **two places**:
1. **Individual resources** — via existing permission methods (CanRead/CanWrite/CanCreate/CanDelete)
2. **Collection queries (ReadAll)** — via a `ProjectScopeable` interface that filters queries

### Key Components

#### 1. `ProjectScoped` Interface (`pkg/web/web.go`)

```go
type ProjectScoped interface {
    GetProjectScope() (projectID int64, includeSubProjects bool)
}
```

Defined alongside the existing `Auth` interface. Any auth object can optionally implement this.

#### 2. User Struct Changes (`pkg/user/user.go`)

Add transient (non-DB) fields:

```go
APITokenProjectID          int64 `xorm:"-" json:"-"`
APITokenIncludeSubProjects bool  `xorm:"-" json:"-"`
```

Implement the `ProjectScoped` interface:

```go
func (u *User) GetProjectScope() (int64, bool) {
    return u.APITokenProjectID, u.APITokenIncludeSubProjects
}
```

#### 3. APIToken Model Changes (`pkg/models/api_tokens.go`)

Add two fields + DB migration:

```go
ProjectID          int64 `xorm:"bigint null" json:"project_id"`
IncludeSubProjects bool  `xorm:"bool default false" json:"include_sub_projects"`
```

Validation in `Create()`: if `ProjectID` is set, verify it exists and the token owner has at least read access to it.

#### 4. Token Middleware Changes (`pkg/routes/api_tokens.go`)

In `checkAPITokenAndPutItInContext`, after fetching the user:

```go
u.APITokenProjectID = token.ProjectID
u.APITokenIncludeSubProjects = token.IncludeSubProjects
c.Set("api_user", u)
```

No other middleware changes needed.

#### 5. Scope Helper Functions (`pkg/models/project_scope.go`)

New file with:

```go
// GetAllChildProjectIDs returns all descendant project IDs for a given project
func GetAllChildProjectIDs(s *xorm.Session, projectID int64) ([]int64, error)

// IsProjectInScope checks if targetProjectID is within scope of scopeProjectID
// If includeSubProjects is true, checks all descendants; otherwise exact match only
func IsProjectInScope(s *xorm.Session, scopeProjectID int64, includeSubProjects bool, targetProjectID int64) (bool, error)

// GetScopeProjectIDs returns the list of project IDs that are in scope.
// For exact match: returns [scopeProjectID]
// For sub-projects: returns [scopeProjectID, ...childIDs]
func GetScopeProjectIDs(s *xorm.Session, a web.Auth) (projectIDs []int64, hasScope bool, err error)
```

`GetAllChildProjectIDs` can use the existing parent_project_id relationships to walk the tree (iterative BFS or a recursive CTE query).

#### 6. Permission Method Changes

Add a scope check helper that can be called from permission methods:

```go
// checkProjectScope checks if the given project is within the auth's project scope.
// Returns (true, nil) if there's no scope or the project is in scope.
// Returns (false, nil) if the project is out of scope.
func checkProjectScope(s *xorm.Session, a web.Auth, projectID int64) (bool, error) {
    scoped, ok := a.(web.ProjectScoped)
    if !ok {
        return true, nil
    }
    scopeProjectID, includeSubProjects := scoped.GetProjectScope()
    if scopeProjectID == 0 {
        return true, nil
    }
    return IsProjectInScope(s, scopeProjectID, includeSubProjects, projectID)
}
```

Then add calls in:

- **`Project.CanRead`** — after getting the project, call `checkProjectScope(s, a, p.ID)`
- **`Project.CanWrite`** — same pattern
- **`Project.CanCreate`** — check that the parent project (or the project itself for top-level) is in scope
- **`Project.CanDelete`** — same as CanWrite
- **`Task.CanRead`** — delegates to `Project.CanRead`, so handled automatically
- **`Task.CanWrite`** — delegates to `Project.CanWrite`, so handled automatically

Other models that delegate to Project permissions (labels, comments, attachments, etc.) are handled transitively.

#### 7. `ProjectScopeable` Interface for ReadAll

For collection queries that build SQL rather than checking individual permissions:

```go
type ProjectScopeable interface {
    ApplyProjectScope(projectIDs []int64)
}
```

Implemented on:
- **`TaskCollection`** — adds `WHERE project_id IN (?)` to the task query
- **`Project` (ReadAll)** — filters the project list to only include in-scope projects

The generic ReadAll handler (or the ReadAll methods themselves) checks if `web.Auth` implements `ProjectScoped`, resolves the scope to project IDs via `GetScopeProjectIDs()`, and calls `ApplyProjectScope()` before executing the query.

#### 8. Database Migration

New migration file adding two columns to the `api_tokens` table:

```sql
ALTER TABLE api_tokens ADD COLUMN project_id BIGINT NULL;
ALTER TABLE api_tokens ADD COLUMN include_sub_projects TINYINT(1) DEFAULT 0;
```

#### 9. Frontend Changes

Update the API token creation/edit UI to add:
- Project selector (dropdown/autocomplete) — optional field
- "Include sub-projects" checkbox — only shown when a project is selected
- Display the scoped project on the token list view

Files to modify:
- `frontend/src/modelTypes/IApiToken.ts` — add `projectId` and `includeSubProjects`
- `frontend/src/components/user/Settings.vue` or equivalent token management component
- Translation strings in `frontend/src/i18n/lang/en.json`

## Implementation Order

1. Database migration (new columns on `api_tokens`)
2. `ProjectScoped` interface in `pkg/web/web.go`
3. User struct changes + `GetProjectScope()` implementation
4. APIToken model changes (new fields, validation)
5. Scope helper functions (`project_scope.go`)
6. `checkProjectScope` helper + permission method integration
7. Token middleware update (set scope fields on user)
8. `ProjectScopeable` for ReadAll on TaskCollection and Project
9. Tests (unit tests for scope helpers, integration tests for permission enforcement)
10. Frontend UI changes

## What This Avoids

- No separate middleware scope enforcement (no `ResolveProjectIDFromRequest`)
- No route path parsing for project IDs in middleware
- No duplicated permission logic
- Minimal changes to the web framework / generic CRUD handler

## Edge Cases

- **Top-level project creation**: Denied when token is project-scoped (can't create projects outside scope)
- **Moving tasks between projects**: CanWrite on the target project will enforce scope
- **Saved filters**: Should respect scope when resolving (the underlying task query will be scoped)
- **Favorites pseudo-project**: ReadAll on favorites should filter to scoped projects only
- **Token without project scope**: `ProjectID == 0` means no scope (current behavior, unchanged)
- **Deleted/inaccessible scope project**: Token becomes useless (all permission checks fail) — this is fine
