# Research: Fix API Token Permissions System

**Date**: October 23, 2025  
**Feature**: Fix API Token Permissions System  
**Phase**: 0 - Research & Analysis

## Research Questions & Findings

### Q1: What is the current state of route registration across v1 API?

**Investigation Approach**: Audit all v1 route files to identify which use declarative APIRoute pattern vs manual registration.

**Findings**:

Based on codebase analysis:

1. **Declarative Pattern (APIRoute + registerRoutes)**:
   - ✅ `task.go` - TaskRoutes array exists (4 routes: create, read_one, update, delete)
   - ✅ `task_positions.go` - Uses declarative pattern
   - ✅ `bulk_tasks.go` - Uses declarative pattern
   - ✅ `task_assignees.go` - Uses declarative pattern
   - ✅ `bulk_assignees.go` - Uses declarative pattern
   - ✅ `task_relations.go` - Uses declarative pattern
   - ✅ `attachments.go` - Uses declarative pattern (if enabled)
   - ✅ `comments.go` - Uses declarative pattern (if enabled)
   - ✅ `project_teams.go` - Uses declarative pattern
   - ✅ `project_users.go` - Uses declarative pattern
   - ✅ `teams.go` - Uses declarative pattern
   - ✅ `subscriptions.go` - Uses declarative pattern
   - ✅ `notifications.go` - Uses declarative pattern
   - ✅ `label_tasks.go` - Uses declarative pattern
   - ✅ `kanban.go` - Uses declarative pattern

2. **Old Pattern (WebHandler or manual Echo registration)**:
   - ⚠️ Projects, Labels - Still use WebHandler pattern (generic CRUD)
   - ⚠️ Saved Filters - Still use WebHandler pattern
   - ⚠️ Background Images - Manual route registration
   - ⚠️ Migrations - Manual route registration
   - ⚠️ API Tokens - Manual route registration (doesn't need token permissions)

**Key Discovery**: TaskRoutes array DOES include all CRUD routes with proper permission scopes. The issue may be elsewhere in the registration chain.

**Decision**: Investigate whether registerRoutes is being called, and whether there are issues with the route path patterns or timing of registration.

---

### Q2: How does the route registration flow work from route definition to GET /routes endpoint?

**Investigation Approach**: Trace the execution path from RegisterTasks() call to apiTokenRoutes map population.

**Findings**:

**Registration Flow**:
```
1. Application Startup (main.go)
   ↓
2. routes.RegisterRoutes(e) (pkg/routes/routes.go)
   ↓
3. registerAPIRoutes(a) - registers v1 routes
   ↓
4. apiv1.RegisterTasks(a) (pkg/routes/api/v1/task.go:48)
   ↓
5. registerRoutes(a, TaskRoutes) (pkg/routes/api/v1/common.go:36)
   ↓
6. For each route in TaskRoutes:
   a. a.Add(route.Method, route.Path, route.Handler) - Registers with Echo
   b. models.CollectRoute(route.Method, route.Path, route.PermissionScope) - Registers permission
   ↓
7. models.CollectRoute (pkg/models/api_routes.go:79)
   - Extracts API version from path (v1, v2)
   - Calls getRouteGroupName(path) to determine group (e.g., "tasks")
   - Stores permission in apiTokenRoutes[version][group][permissionScope]
   ↓
8. GET /routes handler (pkg/models/api_routes.go:389)
   - Returns apiTokenRoutes map as JSON
```

**Key Functions**:
- `getRouteGroupName(path string)` - Maps path to group name
  - `/api/v1/tasks/:taskid` → `"tasks"`
  - `/api/v1/projects/:project/tasks` → `"tasks"` (special case fallthrough)
- `CollectRoute(method, path, permissionScope)` - Stores route permission explicitly

**Verification Method**: Check if all paths in TaskRoutes correctly resolve to "tasks" group.

**Decision**: Test the getRouteGroupName function with actual TaskRoutes paths to verify correct grouping.

---

### Q3: What path patterns exist in TaskRoutes and do they all map to the "tasks" group?

**Investigation Approach**: Extract paths from TaskRoutes and test against getRouteGroupName logic.

**Findings**:

**TaskRoutes Paths**:
1. `PUT /projects/:project/tasks` → Should map to "tasks" (via special case)
2. `GET /tasks/:taskid` → Should map to "tasks"
3. `POST /tasks/:taskid` → Should map to "tasks"
4. `DELETE /tasks/:taskid` → Should map to "tasks"

**getRouteGroupName Logic** (pkg/models/api_routes.go:46):
```go
func getRouteGroupName(path string) (finalName string, filteredParts []string) {
    // Removes /api/v1/ prefix
    // Splits by / and filters out :param parts
    // Joins with _
    
    switch finalName {
    case "projects_tasks":    // Matches PUT /projects/:project/tasks
        fallthrough
    case "tasks_all":
        return "tasks", []string{"tasks"}
    default:
        return finalName, filteredParts
    }
}
```

**Path Analysis**:
- `PUT /projects/:project/tasks` → `"projects_tasks"` → fallthrough → `"tasks"` ✅
- `GET /tasks/:taskid` → `"tasks"` ✅
- `POST /tasks/:taskid` → `"tasks"` ✅
- `DELETE /tasks/:taskid` → `"tasks"` ✅

**Decision**: All TaskRoutes paths correctly map to the "tasks" group. The logic is sound.

---

### Q4: Why might the routes not be appearing in GET /routes response?

**Investigation Approach**: Review the entire CollectRoute function for filtering logic that might exclude certain routes.

**Findings**:

**CollectRoute Filtering** (pkg/models/api_routes.go:79-106):
```go
func CollectRoute(method, path, permissionScope string) {
    routeGroupName, _ := getRouteGroupName(path)
    apiVersion := getRouteAPIVersion(path)
    
    if apiVersion == "" {
        // No api version, no tokens
        return
    }

    // Skip routes that should not be available for API tokens
    if routeGroupName == "tokenTest" ||
        routeGroupName == "subscriptions" ||
        routeGroupName == "tokens" ||
        routeGroupName == "*" ||
        strings.HasPrefix(routeGroupName, "user_") {
        return
    }

    ensureAPITokenRoutesGroup(apiVersion, routeGroupName)
    routeDetail := &RouteDetail{
        Path:   path,
        Method: method,
    }
    apiTokenRoutes[apiVersion][routeGroupName][permissionScope] = routeDetail
}
```

**Key Discovery**: 
- ✅ "tasks" is NOT in the exclusion list
- ✅ Routes with /api/v1/ prefix WILL have apiVersion = "v1"
- ✅ All TaskRoutes should be stored in `apiTokenRoutes["v1"]["tasks"][permissionScope]`

**Potential Issue**: Is registerRoutes actually being called? Or is there a competing registration?

**Decision**: Check routes.go to verify apiv1.RegisterTasks is called and in the correct order.

---

### Q5: Is there a legacy registration system still running that might conflict?

**Investigation Approach**: Search for old CollectRoutesForAPITokenUsage calls or WebHandler registrations for tasks.

**Findings**:

**Legacy Systems**:
1. **CollectRoutesForAPITokenUsage** (pkg/models/api_routes.go:302) - Still exists for backward compatibility
   - Called by Echo after each route is registered via middleware
   - Uses "magic" detection via getRouteDetail function
   - May be overwriting explicit CollectRoute registrations

2. **getRouteDetail** (pkg/models/api_routes.go:109) - Deprecated but still active
   - Has @Deprecated comment
   - Contains fragile pattern matching
   - Returns empty method if pattern doesn't match

**Registration Order**:
```
registerRoutes(a, TaskRoutes)
  → For each route:
      1. a.Add(...) - Registers with Echo
      2. models.CollectRoute(...) - Explicitly registers permission
  
THEN (after a.Add triggers Echo's middleware):
  → CollectRoutesForAPITokenUsage(route, middlewares)
      → Calls getRouteDetail(route)
      → May overwrite or fail to find permission
```

**CRITICAL FINDING**: There are TWO registration systems running:
1. **New System**: Explicit CollectRoute calls during registerRoutes
2. **Legacy System**: CollectRoutesForAPITokenUsage called by Echo middleware

The legacy system may be overwriting the explicit registrations!

**Decision**: The fix is likely to ensure CollectRoute registrations take precedence, OR disable the legacy system for routes using the new pattern.

---

### Q6: How does v2 API registration work compared to v1?

**Investigation Approach**: Examine v2 route files to see if they use CollectRoute or rely on legacy detection.

**Findings**:

**V2 Registration Pattern** (pkg/routes/api/v2/tasks.go):
```go
func RegisterTasks(a *echo.Group) {
    a.GET("/tasks", GetTasks)
}
```

**Analysis**:
- ❌ NO APIRoute array
- ❌ NO CollectRoute calls
- ❌ Relies ENTIRELY on legacy CollectRoutesForAPITokenUsage
- ⚠️ Uses deprecated getRouteDetail "magic" detection

**Impact**: V2 routes depend on the fragile pattern matching in getRouteDetail. If that logic doesn't recognize a route pattern, permissions won't be registered.

**Decision**: V2 routes should ALSO be converted to the declarative pattern for consistency, OR explicit CollectRoute calls should be added.

---

## Best Practices Research

### Best Practice 1: Explicit Route Permission Registration

**Source**: Service layer refactoring patterns (specs/001-complete-service-layer/)

**Pattern**:
```go
type APIRoute struct {
    Method          string
    Path            string
    Handler         echo.HandlerFunc
    PermissionScope string  // Explicit - no guessing
}

var TaskRoutes = []APIRoute{
    {Method: "PUT", Path: "/projects/:project/tasks", Handler: handler, PermissionScope: "create"},
    {Method: "GET", Path: "/tasks/:taskid", Handler: handler, PermissionScope: "read_one"},
    // ... etc
}

func RegisterTasks(a *echo.Group) {
    registerRoutes(a, TaskRoutes)
}
```

**Rationale**: 
- Eliminates fragile pattern matching
- Self-documenting - permission is visible at definition site
- Compile-time safe - typos caught by Go compiler
- Test-friendly - can verify route arrays without running server

**Application**: Already implemented for v1 task routes. Need to verify it's working correctly.

---

### Best Practice 2: Avoid "Magic" Detection Patterns

**Anti-Pattern**: Using reflection, string parsing, or pattern matching to infer behavior

**Example (current codebase - DEPRECATED)**:
```go
func getRouteDetail(route echo.Route) (method string, detail *RouteDetail) {
    if strings.Contains(route.Name, "CreateWeb") {
        return "create", &RouteDetail{...}
    }
    if strings.Contains(route.Name, "UpdateWeb") {
        return "update", &RouteDetail{...}
    }
    // ... lots of fragile string matching
}
```

**Problems**:
- Breaks when function names change
- Requires maintaining complex matching logic
- Hard to debug when patterns don't match
- No compile-time verification

**Recommended Approach**: Explicit declarations (already done with APIRoute pattern)

**Application**: Ensure the deprecated getRouteDetail doesn't override explicit CollectRoute calls.

---

### Best Practice 3: Route Registration Order and Timing

**Pattern**: Register permissions AFTER route is added to router, but BEFORE server starts accepting requests.

**Current Implementation**:
```go
func registerRoutes(a *echo.Group, routes []APIRoute) {
    for _, route := range routes {
        a.Add(route.Method, route.Path, route.Handler)      // Step 1: Register with Echo
        models.CollectRoute(route.Method, route.Path, route.PermissionScope)  // Step 2: Register permission
    }
}
```

**Potential Issue**: If CollectRoutesForAPITokenUsage runs AFTER CollectRoute and doesn't find a match, it might not preserve the explicit registration.

**Solution**: Check CollectRoutesForAPITokenUsage implementation to see how it handles routes that are already registered.

---

## Integration Patterns

### Integration 1: Echo Router + Permission System

**Current Integration**:
```
Echo Router (a.Add)
    ↓
Route registered with Echo
    ↓
Echo middleware calls CollectRoutesForAPITokenUsage (legacy)
    ↓
Explicit CollectRoute call (new)
    ↓
Both write to apiTokenRoutes map
```

**Problem**: Two writers to same data structure without coordination.

**Recommended Pattern**:
```
1. Check if route already has explicit permission in apiTokenRoutes
2. If yes, skip legacy detection
3. If no, attempt legacy detection
```

**Implementation Location**: CollectRoutesForAPITokenUsage function should check if permission already exists before attempting getRouteDetail.

---

### Integration 2: Frontend Permission Display

**Current Implementation** (frontend/src/views/user/settings/ApiTokens.vue):
```vue
<script>
async mounted() {
    this.routes = await this.$services.apiToken.getAvailableRoutes()
    // routes structure: { v1: { tasks: { create: {...}, update: {...} } } }
}
</script>

<template>
    <div v-for="(routes, group) in routes[apiVersion]" :key="group">
        <FancyCheckbox v-model="newTokenPermissions[group][route]">
            {{ formatPermissionTitle(route) }}
        </FancyCheckbox>
    </div>
</template>
```

**Integration Points**:
- GET /routes endpoint returns `apiTokenRoutes` map
- Frontend displays checkboxes for each permission scope
- User selections stored in `IApiToken.permissions` object

**No Changes Needed**: Frontend already correctly displays whatever permissions are returned by GET /routes.

---

## Consolidated Decisions

### Decision 1: Root Cause Identification

**Root Cause**: The legacy CollectRoutesForAPITokenUsage system is likely running AFTER explicit CollectRoute calls and either:
1. Not preserving explicit registrations, OR
2. Not being called at all (middleware not attached), OR
3. Failing to detect routes and leaving gaps

**Evidence**:
- TaskRoutes array exists with all CRUD operations ✅
- registerRoutes is called for tasks ✅
- CollectRoute function is called for each TaskRoute ✅
- But GET /routes is missing some permissions ❌

**Hypothesis**: The issue is in the interaction between the new explicit system and the legacy detection system.

---

### Decision 2: Fix Strategy

**Approach**: Defensive registration that ensures explicit CollectRoute calls take precedence.

**Implementation**:
1. Modify CollectRoutesForAPITokenUsage to check if permission already exists before attempting detection
2. Add logging to CollectRoute to verify it's being called
3. Add logging to CanDoAPIRoute to show which permissions are checked
4. Write tests that verify GET /routes returns expected permissions

**Rationale**: This preserves backward compatibility with routes still using the legacy system, while ensuring the new declarative pattern works correctly.

---

### Decision 3: V2 API Consistency

**Decision**: Convert v2 routes to declarative APIRoute pattern for consistency.

**Rationale**:
- Eliminates dependency on fragile getRouteDetail magic
- Makes v1 and v2 permission registration consistent
- Reduces maintenance burden (one system instead of two)
- Improves testability

**Scope**: Medium effort (3-5 route files in v2), high value (eliminates legacy technical debt).

**Alternatives Considered**:
- Keep v2 with legacy system: Rejected - perpetuates technical debt
- Add explicit CollectRoute calls without APIRoute: Rejected - inconsistent with v1 pattern

---

### Decision 4: Testing Strategy

**Test Levels**:
1. **Unit Tests**: Test CollectRoute, getRouteGroupName, CanDoAPIRoute functions
2. **Integration Tests**: Test that registerRoutes correctly populates apiTokenRoutes
3. **E2E Tests**: Test GET /routes endpoint returns complete permissions
4. **HTTP Tests**: Test actual API token authentication for create/update/delete operations

**Test Files**:
- `pkg/models/api_routes_test.go` - Unit tests for route registration logic
- `pkg/services/api_tokens_test.go` - Integration tests for token validation
- `pkg/webtests/api_tokens_test.go` - HTTP-level tests for token authentication

---

## Summary

**Key Findings**:
1. ✅ Declarative APIRoute pattern is correctly implemented for v1 tasks
2. ✅ CollectRoute function correctly stores explicit permissions
3. ⚠️ Legacy CollectRoutesForAPITokenUsage may be interfering with explicit registrations
4. ❌ V2 routes rely entirely on deprecated getRouteDetail magic

**Root Cause**: Conflict between explicit permission registration and legacy detection system.

**Fix Approach**:
1. Make explicit CollectRoute calls take precedence over legacy detection
2. Add defensive checks in CollectRoutesForAPITokenUsage
3. Convert v2 routes to declarative pattern
4. Add comprehensive tests to prevent regression

**Alternatives Considered**:
- Remove legacy system entirely: Rejected - too risky, may break routes not yet converted
- Only fix v1: Rejected - v2 would remain inconsistent
- Revert to all legacy: Rejected - violates architecture principles

**Next Phase**: Design data model and API contracts (Phase 1).
