# Feature Specification: Fix API Token Permissions System

**Feature Branch**: `005-fix-api-token-permissions`  
**Created**: October 23, 2025  
**Status**: Draft  
**Input**: User description: "Fix API token permissions system to ensure all v1 and v2 API routes are properly registered with their permission scopes for token-based authentication"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - API Token Task Management (Priority: P1)

A user creates an API token with "all permissions" selected and attempts to create, update, and delete tasks using the v1 API. Currently, these operations fail with authentication errors despite the token having "all permissions" because the create, update, and delete permission scopes are not registered in the route system.

**Why this priority**: This is the core functionality that is broken. Users cannot perform basic CRUD operations on tasks with API tokens, which is a critical feature regression. This blocks API integrations and automation workflows.

**Independent Test**: Create an API token with "v1_tasks" permissions including create/update/delete, then attempt to create a task via PUT /api/v1/projects/:project/tasks, update a task via POST /api/v1/tasks/:taskid, and delete a task via DELETE /api/v1/tasks/:taskid. All operations should succeed with proper authorization.

**Acceptance Scenarios**:

1. **Given** a user has created an API token with "all permissions" selected, **When** they make a PUT request to /api/v1/projects/:project/tasks with valid task data, **Then** the task is created successfully and returns HTTP 201
2. **Given** a user has created an API token with v1_tasks create/update/delete permissions, **When** they make a POST request to /api/v1/tasks/:taskid with updated task data, **Then** the task is updated successfully and returns HTTP 200
3. **Given** a user has created an API token with v1_tasks delete permission, **When** they make a DELETE request to /api/v1/tasks/:taskid, **Then** the task is deleted successfully and returns HTTP 200
4. **Given** a user has created an API token without v1_tasks create permission, **When** they attempt to create a task, **Then** the request is rejected with HTTP 401 Unauthorized

---

### User Story 2 - Complete Permission Scope Discovery (Priority: P1)

A user navigates to the API token creation page in the frontend UI to view all available permissions. Currently, the GET /routes endpoint returns incomplete permission scopes for v1_tasks (missing create, update, delete), preventing users from selecting these permissions during token creation.

**Why this priority**: Users cannot create properly scoped tokens if the UI doesn't show all available permissions. This is a prerequisite for User Story 1 to work correctly from the UI perspective.

**Independent Test**: Call GET /routes endpoint and verify that v1_tasks group includes create, update, delete, and read_one permission scopes. Frontend UI should display checkboxes for all four operations under the tasks section.

**Acceptance Scenarios**:

1. **Given** a user is authenticated, **When** they call GET /api/v1/routes, **Then** the response includes v1_tasks with create, update, delete, and read_one permissions
2. **Given** a user navigates to the API token creation page, **When** the page loads, **Then** the tasks section displays checkboxes for create, update, delete, and read operations
3. **Given** a user checks "select all permissions" checkbox, **When** the token is created, **Then** all CRUD operations are included in the token's permission set
4. **Given** a user selects only specific task permissions (e.g., read and create), **When** the token is created, **Then** only those specific permissions are granted

---

### User Story 3 - V2 API Route Consistency (Priority: P2)

A user creates an API token for v2 API routes and verifies that all v2 routes are properly registered with their permission scopes, ensuring consistency between v1 and v2 API behavior for token-based authentication.

**Why this priority**: While v1 routes are the immediate issue, v2 API consistency is important for future-proofing and maintaining a coherent API token system across all API versions.

**Independent Test**: Call GET /routes endpoint and verify that v2 routes (v2_tasks, v2_projects, v2_labels) are registered with complete CRUD permission scopes. Test token authentication against v2 API endpoints.

**Acceptance Scenarios**:

1. **Given** a user calls GET /api/v1/routes, **When** the response is returned, **Then** v2 route groups include complete CRUD permission scopes
2. **Given** a user creates an API token with v2_tasks permissions, **When** they perform CRUD operations via v2 API endpoints, **Then** all operations succeed with proper authorization
3. **Given** a user has tokens for both v1 and v2 APIs, **When** comparing permission structures, **Then** both versions follow the same permission naming conventions and scope patterns

---

### User Story 4 - Existing Token Backward Compatibility (Priority: P1)

A user with existing API tokens that were created before the fix continues to use those tokens without disruption. The system maintains backward compatibility with tokens that may have permissions configured using the old registration system.

**Why this priority**: Breaking existing API tokens would cause production failures for users who have already integrated with the API. Backward compatibility is critical for a smooth transition.

**Independent Test**: Use a token created before the fix (if possible via database seeding or migration) and verify it still works for routes it had permission to access. Create new tokens and verify they work with the fixed permission system.

**Acceptance Scenarios**:

1. **Given** a user has an API token created before the permission system fix, **When** they use that token to access routes they previously had permission for, **Then** the token continues to work without errors
2. **Given** a user has an old token with "v1_tasks" read permission (without version prefix), **When** they attempt to read tasks, **Then** the system checks both versioned (v1_tasks) and non-versioned (tasks) permission keys for backward compatibility
3. **Given** a user creates a new token after the fix, **When** they select permissions from the updated UI, **Then** the token uses the new permission structure and works correctly

---

### Edge Cases

- What happens when a route is registered multiple times with different permission scopes?
- How does the system handle routes that have both old "magic" detection and new explicit registration?
- What happens when a token has permissions for a route group that no longer exists after refactoring?
- How does the system handle routes with dynamic path parameters (e.g., /api/v1/tasks/:taskid)?
- What happens if GET /routes is called before any routes are registered during application startup?
- How does the system handle routes that are conditionally registered based on configuration flags (e.g., attachments, comments)?
- What happens when a token has overlapping permissions (e.g., both "tasks" and "v1_tasks")?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST register all v1 task routes (create, read_one, update, delete) with their correct permission scopes in the API token route map
- **FR-002**: System MUST expose all registered route permissions through the GET /routes endpoint for token permission selection
- **FR-003**: System MUST validate API token permissions against registered route scopes using the CanDoAPIRoute function
- **FR-004**: System MUST maintain backward compatibility with existing API tokens that use non-versioned permission keys (e.g., "tasks" instead of "v1_tasks")
- **FR-005**: System MUST register all v2 API routes with their correct permission scopes consistently with v1 route registration
- **FR-006**: System MUST prevent duplicate or conflicting permission scope registrations for the same route
- **FR-007**: System MUST ensure that the registerRoutes helper function correctly calls both Echo route registration and CollectRoute permission registration
- **FR-008**: System MUST support all HTTP methods (GET, POST, PUT, DELETE, PATCH) for route permission registration
- **FR-009**: Frontend MUST display all available permission scopes from GET /routes endpoint in the token creation UI
- **FR-010**: Frontend "select all permissions" checkbox MUST include all CRUD operations for all resource types returned by GET /routes
- **FR-011**: System MUST log authentication failures when a token lacks required permissions, including the missing permission scope
- **FR-012**: System MUST handle routes with path parameters correctly in permission checking (e.g., :taskid, :project)

### Key Entities *(include if feature involves data)*

- **APIToken**: Represents an API token with permissions field containing a map of versioned route groups to permission scopes (e.g., {"v1_tasks": ["create", "update", "delete", "read_one"]})
- **APITokenRoute**: A map structure storing route groups and their available permission scopes, indexed by API version
- **RouteDetail**: Contains the HTTP path and method for a registered route
- **APIRoute**: Structure defining a route registration with Method, Path, Handler, and PermissionScope fields

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: API tokens with v1_tasks create permission can successfully create tasks via PUT /api/v1/projects/:project/tasks (100% success rate)
- **SC-002**: API tokens with v1_tasks update permission can successfully update tasks via POST /api/v1/tasks/:taskid (100% success rate)
- **SC-003**: API tokens with v1_tasks delete permission can successfully delete tasks via DELETE /api/v1/tasks/:taskid (100% success rate)
- **SC-004**: GET /routes endpoint returns complete permission scopes including create, update, delete for v1_tasks group (verified by automated test)
- **SC-005**: Frontend token creation UI displays all CRUD operations (create, read, update, delete) for tasks resource type (verified by E2E test)
- **SC-006**: "Select all permissions" checkbox in frontend includes all available permission scopes from GET /routes (verified by E2E test)
- **SC-007**: Existing API tokens created before the fix continue to work for routes they had permission to access (zero breaking changes)
- **SC-008**: All v2 API routes are registered with complete permission scopes matching v1 route patterns (verified by automated test)
- **SC-009**: API authentication logs include specific permission scope information for debugging unauthorized requests (100% of auth failures logged with permission details)
- **SC-010**: No regression in existing API token functionality for routes that were already working (all existing passing tests continue to pass)

