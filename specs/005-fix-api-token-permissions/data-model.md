# Data Model: Fix API Token Permissions System

**Date**: October 23, 2025  
**Feature**: Fix API Token Permissions System  
**Phase**: 1 - Design & Contracts

## Overview

This fix does NOT introduce new data models. It corrects the registration and validation of existing data structures used by the API token permission system. This document describes the relevant existing entities and their relationships.

## Existing Entities

### 1. APIToken

**Purpose**: Represents an API authentication token with scoped permissions.

**Location**: `pkg/models/api_tokens.go`

**Fields**:
| Field | Type | Description | Validation Rules |
|-------|------|-------------|------------------|
| ID | int64 | Unique token identifier | Auto-increment, primary key |
| Title | string | Human-readable token name | Required, user-defined |
| Token | string | Actual token string (visible only on creation) | Generated, prefixed with `tk_` |
| TokenHash | string | Secure hash of token | PBKDF2 hash, stored in DB |
| TokenSalt | string | Salt for token hash | Random bytes, stored in DB |
| APIPermissions | APIPermissions | Map of route groups to permission scopes | Required, validated against apiTokenRoutes |
| ExpiresAt | time.Time | Token expiration timestamp | Required, future date |
| Created | time.Time | Creation timestamp | Auto-set, immutable |
| OwnerID | int64 | User who owns this token | Foreign key to users table |

**Relationships**:
- Belongs to one User (OwnerID)
- Has many Permissions (embedded in APIPermissions field)

**State Transitions**:
```
[Created] → [Active] → [Expired]
                ↓
           [Deleted]
```

**Validation Rules**:
- APIPermissions MUST only contain valid route groups and permission scopes from `apiTokenRoutes`
- ExpiresAt MUST be in the future at creation time
- Title MUST be non-empty
- Token MUST be unique across all tokens

**No Schema Changes Required**: This fix does not modify the APIToken table structure.

---

### 2. APIPermissions

**Purpose**: Type alias for a map of versioned route groups to permission scope arrays.

**Location**: `pkg/models/api_tokens.go`

**Structure**:
```go
type APIPermissions map[string][]string
```

**Example Values**:
```json
{
  "v1_tasks": ["create", "update", "delete", "read_one"],
  "v1_projects": ["create", "read_all", "read_one", "update", "delete"],
  "v2_tasks": ["read_all"]
}
```

**Key Format**: `{version}_{groupName}` (e.g., "v1_tasks", "v2_projects")

**Value Format**: Array of permission scope strings (e.g., ["create", "update", "delete"])

**Validation Rules**:
- Keys MUST match pattern: `v[0-9]+_[a-z_]+`
- Keys MUST exist in `apiTokenRoutes[version][groupName]`
- Values MUST be arrays of strings
- Each string in values MUST exist in `apiTokenRoutes[version][groupName]` as a permission scope

**Backward Compatibility**:
- Also accepts non-versioned keys (e.g., "tasks") for old tokens
- CanDoAPIRoute checks both versioned and non-versioned keys

---

### 3. APITokenRoute

**Purpose**: Type alias for a map of permission scopes to route details.

**Location**: `pkg/models/api_routes.go`

**Structure**:
```go
type APITokenRoute map[string]*RouteDetail
```

**Example Value**:
```go
map[string]*RouteDetail{
    "create":   &RouteDetail{Path: "/api/v1/projects/:project/tasks", Method: "PUT"},
    "read_one": &RouteDetail{Path: "/api/v1/tasks/:taskid", Method: "GET"},
    "update":   &RouteDetail{Path: "/api/v1/tasks/:taskid", Method: "POST"},
    "delete":   &RouteDetail{Path: "/api/v1/tasks/:taskid", Method: "DELETE"},
}
```

**Keys**: Permission scope strings (e.g., "create", "read_all", "update")

**Values**: Pointers to RouteDetail structs

**Storage**: In-memory map `apiTokenRoutes` with structure:
```
apiTokenRoutes[version][groupName][permissionScope] = *RouteDetail
```

**Population**:
- Explicit registration via `CollectRoute(method, path, permissionScope)`
- Legacy registration via `CollectRoutesForAPITokenUsage(route, middlewares)` (deprecated)

---

### 4. RouteDetail

**Purpose**: Stores HTTP method and path for a registered route.

**Location**: `pkg/models/api_routes.go`

**Structure**:
```go
type RouteDetail struct {
    Path   string `json:"path"`
    Method string `json:"method"`
}
```

**Fields**:
| Field | Type | Description |
|-------|------|-------------|
| Path | string | HTTP path pattern (e.g., "/api/v1/tasks/:taskid") |
| Method | string | HTTP method (GET, POST, PUT, DELETE, PATCH) |

**Usage**:
- Stored in `apiTokenRoutes` map
- Returned by GET /routes endpoint for frontend display
- Used by CanDoAPIRoute for permission checking

---

### 5. APIRoute (Declarative Pattern)

**Purpose**: Defines a single API endpoint with explicit permission scope (new pattern).

**Location**: `pkg/routes/api/v1/common.go`

**Structure**:
```go
type APIRoute struct {
    Method          string
    Path            string
    Handler         echo.HandlerFunc
    PermissionScope string
}
```

**Fields**:
| Field | Type | Description |
|-------|------|-------------|
| Method | string | HTTP method (GET, POST, PUT, DELETE, PATCH) |
| Path | string | HTTP path pattern (e.g., "/tasks/:taskid") |
| Handler | echo.HandlerFunc | Request handler function |
| PermissionScope | string | Explicit permission name (e.g., "create", "update") |

**Usage**:
- Defined in route arrays (e.g., `var TaskRoutes = []APIRoute{...}`)
- Processed by `registerRoutes(a, routes)` helper
- Each route triggers both Echo registration AND CollectRoute call

**Benefits**:
- Self-documenting (permission visible at definition)
- Compile-time safe (typos caught by compiler)
- No "magic" detection required

---

## Data Flow Diagrams

### Route Registration Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ Application Startup                                             │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ RegisterRoutes(e) in routes.go                                  │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ registerAPIRoutes(a) - v1 routes                                │
│ registerAPIRoutesV2(a) - v2 routes                              │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ For v1: apiv1.RegisterTasks(a)                                  │
│ For v2: apiv2.RegisterTasks(a)                                  │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ├─────────── v1 (Declarative) ─────────────┐
                   │                                           │
                   ▼                                           │
   ┌───────────────────────────────────────┐                  │
   │ registerRoutes(a, TaskRoutes)         │                  │
   │ for each route:                       │                  │
   │   1. a.Add(method, path, handler)     │                  │
   │   2. models.CollectRoute(...)         │                  │
   └───────────────┬───────────────────────┘                  │
                   │                                           │
                   ├─────────── v2 (Manual) ──────────────────┘
                   │
                   ▼
   ┌───────────────────────────────────────┐
   │ a.GET("/tasks", GetTasks)             │
   │ (direct Echo registration)            │
   └───────────────┬───────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ apiTokenRoutes[version][group][scope] = RouteDetail            │
│ In-memory map populated                                         │
└─────────────────────────────────────────────────────────────────┘
```

### Permission Validation Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ HTTP Request with API Token                                     │
│ Authorization: Bearer tk_xxxxx...                               │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ checkAPITokenAndPutItInContext(tokenHeaderValue, c)            │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ tokenService.GetTokenFromTokenString(s, token)                 │
│ → Validates hash, checks expiration                             │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ models.CanDoAPIRoute(c, token)                                  │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ Extract: path, method from request                              │
│ Extract: apiVersion, routeGroupName from path                   │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ Check token.APIPermissions[version_group]                       │
│ (Also check backward compat: token.APIPermissions[group])      │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ Lookup apiTokenRoutes[version][group] for available scopes     │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ Match request path+method against RouteDetail                   │
│ Check if token has required permission scope                    │
└──────────────────┬──────────────────────────────────────────────┘
                   │
       ┌───────────┴────────────┐
       │                        │
       ▼                        ▼
   [ALLOW]                 [DENY]
   Return true            Return false
   Request proceeds       401 Unauthorized
```

### GET /routes Endpoint Flow

```
┌─────────────────────────────────────────────────────────────────┐
│ GET /api/v1/routes                                              │
│ (Authenticated user)                                            │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ models.GetAvailableAPIRoutesForToken(c)                        │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ Return apiTokenRoutes map as JSON                               │
│ {                                                                │
│   "v1": {                                                        │
│     "tasks": {                                                   │
│       "create": {"path": "...", "method": "PUT"},               │
│       "update": {"path": "...", "method": "POST"},              │
│       "delete": {"path": "...", "method": "DELETE"}             │
│     }                                                            │
│   }                                                              │
│ }                                                                │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│ Frontend ApiTokens.vue                                          │
│ Displays checkboxes for each permission scope                   │
└─────────────────────────────────────────────────────────────────┘
```

## Entity Relationships

```
┌─────────────────┐
│     User        │
└────────┬────────┘
         │ 1
         │
         │ owns
         │
         │ *
┌────────▼────────┐
│   APIToken      │
│                 │
│ APIPermissions  ├──────────┐
└─────────────────┘          │ validates against
                              │
                              ▼
                    ┌─────────────────────┐
                    │  apiTokenRoutes     │
                    │  (in-memory map)    │
                    │                     │
                    │  [version]          │
                    │    [groupName]      │
                    │      [scope] ────┐  │
                    └──────────────────┼──┘
                                       │
                                       │ *
                                       ▼
                              ┌─────────────────┐
                              │  RouteDetail    │
                              │                 │
                              │  Path: string   │
                              │  Method: string │
                              └─────────────────┘
```

## Validation Rules Summary

### APIToken Creation
1. Title MUST be non-empty
2. ExpiresAt MUST be future date
3. APIPermissions MUST pass PermissionsAreValid check:
   - Each key MUST be `{version}_{group}` format
   - Each `{version}` MUST exist in apiTokenRoutes
   - Each `{group}` MUST exist in apiTokenRoutes[version]
   - Each permission scope in values array MUST exist in apiTokenRoutes[version][group]

### Permission Checking (CanDoAPIRoute)
1. Extract version and group from request path
2. Check token.APIPermissions[version_group] exists (try non-versioned as fallback)
3. Lookup apiTokenRoutes[version][group] for available routes
4. Match request path+method against RouteDetail
5. Verify token has required permission scope

### Route Registration (CollectRoute)
1. Extract version from path (must be /api/v{number}/)
2. Extract group name from path (filter out :params)
3. Skip if group is in exclusion list (tokenTest, subscriptions, tokens, user_*)
4. Store in apiTokenRoutes[version][group][permissionScope] = RouteDetail

## No Schema Migrations Required

This fix does NOT modify database schemas. All changes are to in-memory registration and validation logic only.

**Affected Tables**: None  
**New Tables**: None  
**Modified Columns**: None

The APIToken table schema remains unchanged. The fix ensures the existing schema is correctly validated against properly registered routes.
