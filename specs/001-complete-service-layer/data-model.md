# Data Model: Complete Service-Layer Refactor Stabilization and Validation

## Core Architectural Entities

### Service Layer Architecture
**Purpose**: Three-layer architecture implementation
**Fields**:
- `Services`: Business logic layer (pkg/services) - contains ALL business logic
- `Handlers`: Request/response layer (pkg/routes) - thin glue using handler wrappers  
- `Models`: Data access layer (pkg/models) - CRUD operations only, no business logic

**Relationships**:
- Handlers → Services (one-way dependency)
- Services → Models (one-way dependency)  
- Models ↔ Services (dependency inversion via function variables)

**Validation Rules**:
- Services MUST contain all business logic
- Handlers MUST be thin and use declarative routing
- Models MUST NOT contain business logic
- All layers MUST have >90% test coverage for refactored components

**State Transitions**:
- Current State: Partially refactored (some services exist, some business logic remains in models)
- Target State: Fully refactored (all business logic in services, models are data-only)

### Route Architecture
**Purpose**: Consistent routing pattern across entire application
**Fields**:
- `APIRoute`: Declarative route definition with explicit permissions
- `WebHandler`: Legacy route pattern with implicit permissions (deprecated)
- `RegisterFunction`: Modular route registration in pkg/routes/api/v1/

**Relationships**:
- APIRoute → handler.WithDBAndUser() wrapper
- RegisterFunction → registerRoutes() helper
- routes.go → Multiple RegisterFunction calls

**Validation Rules**:
- All routes MUST use APIRoute declarative pattern
- All routes MUST declare explicit permission scopes
- All route files MUST be in pkg/routes/api/v1/ directory
- routes.go MUST contain only framework setup and Register* calls (<250 lines)

**State Transitions**:
- Current State: Mixed patterns (9 modern, 15 legacy WebHandler)
- Target State: 100% modern declarative pattern (24+ route files)

**Migration Pattern Example**:
```go
// Legacy (routes.go)
handler := &handler.WebHandler{EmptyStruct: func() {return &models.Entity{}}}
a.PUT("/path", handler.CreateWeb)

// Modern (pkg/routes/api/v1/entity.go)
var EntityRoutes = []APIRoute{
    {Method: "PUT", Path: "/path", Handler: handler.WithDBAndUser(createLogic, true), PermissionScope: "create"},
}
func RegisterEntity(a *echo.Group) { registerRoutes(a, EntityRoutes) }
```

### Test Infrastructure  
**Purpose**: Comprehensive testing across all layers
**Fields**:
- `ModelTests`: Simple unit tests, decoupled from services
- `ServiceTests`: Integration tests with full dependency injection
- `HandlerTests`: API endpoint tests using handler wrappers
- `ParityTests`: Comparison tests between original and refactored systems

**Relationships**:
- ServiceTests → testutil.Init() (dependency injection)
- ModelTests → Mock functions (no service dependencies)
- ParityTests → Both original and refactored implementations

**Validation Rules**:
- Model tests MUST NOT call testutil.Init()
- Service tests MUST call testutil.Init() for full wiring
- All tests MUST follow TDD (test-first) approach
- Parity tests MUST validate identical behavior

### Validation Framework
**Purpose**: Systematic comparison and verification processes  
**Fields**:
- `TestParityAnalysis`: Automated comparison of test suites
- `FunctionalParityChecklist`: Manual validation workflows
- `ArchitecturalReview`: AI analysis + human approval process

**Relationships**:
- TestParityAnalysis → vikunja/ and vikunja_original_main/ test suites
- FunctionalParityChecklist → Core user workflows
- ArchitecturalReview → Complete codebase analysis

**Validation Rules**:
- Original system behavior takes precedence when differences found
- All test cases from original system MUST be preserved
- Manual validation MUST cover all core user workflows

### Refactor Backlog
**Purpose**: Cataloged features requiring service-layer implementation
**Fields**:
- `FeatureName`: Name of feature to refactor
- `ModelFiles`: Associated model files containing business logic
- `Complexity`: High/Medium/Low complexity estimate
- `Dependencies`: Other features that must be refactored first
- `Priority`: Calculated from dependency + complexity

**High Complexity Features** (Business logic heavy):
- Projects (project.go)
- Tasks (tasks.go)  
- Duplicate Project (project_duplicate.go)
- Saved Filters (saved_filters.go)
- Subscriptions (subscription.go)
- Project Views (project_view.go)
- Link Sharing (link_sharing.go)
- Label-Task Management (label_task.go)
- User Data Export (export.go)
- Kanban Buckets (kanban.go)
- Bulk Task Update (bulk_task.go)

**Medium Complexity Features**:
- API Tokens (api_tokens.go)
- Project-Team Permissions (project_team.go) 
- Project-User Permissions (project_users.go)
- Labels (label.go)
- Reactions (reaction.go)
- Notifications (notifications.go, notifications_database.go)

**Low Complexity Features**:
- Favorites (favorites.go)
- User Mentions (mentions.go)

**Dependency Relationships**:
- Labels → Label-Task Management
- Projects → Project Views, Project-Team Permissions, Project-User Permissions
- Tasks → Bulk Task Update, Kanban Buckets
- Users → API Tokens, Notifications, User Mentions

### Quality Gates
**Purpose**: Defined checkpoints and criteria for phase completion
**Fields**:
- `TestPassRate`: Percentage of tests passing (target: 100%)
- `CoverageMetrics`: Test coverage percentages by component
- `PerformanceBenchmarks`: API response time measurements  
- `ArchitecturalCompliance`: Adherence to constitutional principles

**Phase 1 Gates**:
- All backend tests pass (mage test:feature)
- Task-related query failures resolved
- UI bugs fixed (404 on label creation, empty task detail view)
- Functional parity with vikunja_original_main demonstrated

**Phase 2 Gates**:
- All 18 features refactored following dependency → complexity order
- Service layer implements "Chef, Waiter, Pantry" pattern
- Declarative routing with handler wrappers implemented
- Dependency inversion pattern applied for backward compatibility
- All legacy WebHandler routes migrated to modern APIRoute pattern
- routes.go cleaned up to <250 lines (framework setup only)
- 100% architectural consistency achieved

**Phase 3 Gates**:
- Automated test parity analysis confirms no lost test cases
- Functional parity checklist executed and validated
- Architectural review passed (AI analysis + human approval)
- 90% test coverage achieved for refactored service layer components
- Performance requirements maintained (<200ms API response time)

## Data Flow Patterns

### Request Processing Flow
```
HTTP Request → Handler (thin glue) → Service (business logic) → Model (data access) → Database
HTTP Response ← Handler (response formatting) ← Service (business result) ← Model (data) ← Database
```

### Test-Driven Development Flow
```
Write Test (fail) → Implement Service Logic → Test Passes → Refactor → Update Handlers → Integration Test
```

### Dependency Inversion Flow
```
Model Function Variable ← Service Implementation (via init()) → Handler Calls Service → Service Updates Model Variable
```

This data model ensures complete architectural transformation while maintaining functional parity with the original system.