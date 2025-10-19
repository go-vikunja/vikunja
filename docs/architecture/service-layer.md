# Service Layer Architecture

This document provides a visual overview of Vikunja's three-layer service architecture, showing how requests flow through the system and where different concerns are handled.

## Architecture Overview

Vikunja follows a clean three-layer architecture pattern, often referred to as the "Chef, Waiter, Pantry" model:

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Request                              │
│                 (POST /api/v1/tasks/:task)                       │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────────┐
│                      🧑‍💼 HANDLERS LAYER                            │
│                    (pkg/routes/api/v1/)                           │
│─────────────────────────────────────────────────────────────────│
│  Role: "Waiter" - Thin glue layer                                 │
│  - Parse HTTP requests                                            │
│  - Extract parameters & body                                      │
│  - Call appropriate service method                                │
│  - Format HTTP responses                                          │
│  - Set headers (pagination, etc.)                                 │
│                                                                   │
│  Example: updateTaskLogic(s, u, c)                                │
│    1. Parse task ID from URL                                      │
│    2. Bind request body to Task struct                            │
│    3. Call taskService.Update(s, task, u)                         │
│    4. Return JSON response                                        │
└───────────────────────────────┬───────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────────┐
│                       🧑‍🍳 SERVICES LAYER                            │
│                       (pkg/services/)                             │
│─────────────────────────────────────────────────────────────────│
│  Role: "Chef" - ALL business logic                                │
│  - Permission checks (BEFORE existence checks)                    │
│  - Input validation & sanitization                                │
│  - Business rules enforcement                                     │
│  - Transaction management                                         │
│  - Event dispatching                                              │
│  - Orchestration of multiple operations                           │
│                                                                   │
│  Example: TaskService.Update(s, task, u)                          │
│    1. Fetch existing task from database                           │
│    2. Check user has update permission                            │
│    3. Validate input (title not empty, etc.)                      │
│    4. Apply business rules                                        │
│    5. Update database via models                                  │
│    6. Dispatch TaskUpdatedEvent                                   │
│    7. Return updated task                                         │
└───────────────────────────────┬───────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────────┐
│                       🏪 MODELS LAYER                              │
│                       (pkg/models/)                               │
│─────────────────────────────────────────────────────────────────│
│  Role: "Pantry" - Data access only                                │
│  - Database table definitions (structs)                           │
│  - XORM mappings & relationships                                  │
│  - Simple CRUD operations                                         │
│  - NO business logic                                              │
│  - NO permission checks                                           │
│                                                                   │
│  Example: Task struct & basic DB operations                       │
│    - Field definitions (ID, Title, ProjectID, etc.)               │
│    - Table name mapping: TableName() -> "tasks"                   │
│    - Deprecated facades that delegate to services                 │
└───────────────────────────────┬───────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         DATABASE                                  │
│                    (SQLite/PostgreSQL/MySQL)                      │
└─────────────────────────────────────────────────────────────────┘
```

## Request Flow Example: Update Task

Here's how a task update request flows through the layers:

```
1. HTTP Request arrives
   ┌─────────────────────────────────────────────────────┐
   │ PUT /api/v1/tasks/42                                │
   │ Authorization: Bearer <token>                       │
   │ Body: {"title": "Updated Title", "done": true}      │
   └─────────────────────────────────────────────────────┘
                        │
                        ▼
2. Handler Layer (pkg/routes/api/v1/task.go)
   ┌─────────────────────────────────────────────────────┐
   │ updateTaskLogic(s *xorm.Session, u *user.User, c)  │
   │   taskID := c.Param("task") // "42"                 │
   │   task := &models.Task{}                            │
   │   c.Bind(task) // Parse request body                │
   │   task.ID = taskID                                  │
   │                                                     │
   │   service := services.NewTaskService(db)            │
   │   updated, err := service.Update(s, task, u)        │
   │   if err != nil { return err }                      │
   │                                                     │
   │   return c.JSON(200, updated)                       │
   └─────────────────────────────────────────────────────┘
                        │
                        ▼
3. Service Layer (pkg/services/task.go)
   ┌─────────────────────────────────────────────────────┐
   │ TaskService.Update(s, task, u)                      │
   │   // 1. Fetch existing task                         │
   │   existing, err := getTaskByID(s, task.ID)          │
   │   if err != nil {                                   │
   │     return nil, ErrGenericForbidden{} // Hide       │
   │   }                                                 │
   │                                                     │
   │   // 2. Permission check BEFORE revealing existence │
   │   canUpdate, err := checkPermission(s, existing, u) │
   │   if !canUpdate {                                   │
   │     return nil, ErrGenericForbidden{}               │
   │   }                                                 │
   │                                                     │
   │   // 3. Validate business rules                     │
   │   if task.Title == "" {                             │
   │     return nil, ErrTaskTitleCannotBeEmpty{}         │
   │   }                                                 │
   │                                                     │
   │   // 4. Apply updates                               │
   │   existing.Title = task.Title                       │
   │   existing.Done = task.Done                         │
   │   existing.Updated = time.Now()                     │
   │                                                     │
   │   // 5. Persist to database                         │
   │   _, err = s.Where("id = ?", existing.ID).          │
   │              Update(existing)                       │
   │   if err != nil { return nil, err }                 │
   │                                                     │
   │   // 6. Dispatch event                              │
   │   events.Dispatch(&TaskUpdatedEvent{existing})      │
   │                                                     │
   │   return existing, nil                              │
   └─────────────────────────────────────────────────────┘
                        │
                        ▼
4. Models Layer (pkg/models/task.go)
   ┌─────────────────────────────────────────────────────┐
   │ Task struct (data container)                        │
   │   type Task struct {                                │
   │     ID          int64                               │
   │     Title       string                              │
   │     Done        bool                                │
   │     ProjectID   int64                               │
   │     CreatedByID int64                               │
   │     Created     time.Time                           │
   │     Updated     time.Time                           │
   │   }                                                 │
   │                                                     │
   │   func (t *Task) TableName() string {               │
   │     return "tasks"                                  │
   │   }                                                 │
   │                                                     │
   │   // Deprecated facade                              │
   │   func (t *Task) Update(s, u) error {               │
   │     return TaskUpdateFunc(s, t, u) // Delegate      │
   │   }                                                 │
   └─────────────────────────────────────────────────────┘
                        │
                        ▼
5. Database Query
   ┌─────────────────────────────────────────────────────┐
   │ UPDATE tasks                                        │
   │ SET title = 'Updated Title',                        │
   │     done = true,                                    │
   │     updated = '2025-10-08 12:34:56'                 │
   │ WHERE id = 42                                       │
   └─────────────────────────────────────────────────────┘
                        │
                        ▼
6. Response flows back up the layers
   ┌─────────────────────────────────────────────────────┐
   │ HTTP Response                                       │
   │ Status: 200 OK                                      │
   │ Body: {                                             │
   │   "id": 42,                                         │
   │   "title": "Updated Title",                         │
   │   "done": true,                                     │
   │   "created": "2025-01-01T10:00:00Z",                │
   │   "updated": "2025-10-08T12:34:56Z"                 │
   │ }                                                   │
   └─────────────────────────────────────────────────────┘
```

## Permission Checking Flow

One of the most important security patterns in the service layer is **permission-before-existence** checking:

```
┌─────────────────────────────────────────────────────────────────┐
│                 Permission Check Flow                            │
└─────────────────────────────────────────────────────────────────┘

Request: GET /api/v1/tasks/999 (task user doesn't have access to)
                        │
                        ▼
           ┌────────────────────────┐
           │   Handler Layer        │
           │  Extract task ID: 999  │
           └───────────┬────────────┘
                       │
                       ▼
           ┌────────────────────────────────────────┐
           │         Service Layer                  │
           ├────────────────────────────────────────┤
           │ 1. Fetch task from database            │
           │    task, err := getTaskByID(s, 999)    │
           │    if err != nil {                     │
           │      return nil, ErrGenericForbidden{} │ <-- Don't reveal "not found"
           │    }                                   │
           │                                        │
           │ 2. Check permission                    │
           │    hasAccess := checkPermission(task)  │
           │    if !hasAccess {                     │
           │      return nil, ErrGenericForbidden{} │ <-- Same error as "not found"
           │    }                                   │
           │                                        │
           │ 3. Only now reveal task details        │
           │    return task, nil                    │
           └────────────────────────────────────────┘
                       │
                       ▼
           ┌────────────────────────┐
           │   Handler Layer        │
           │  Return 403 Forbidden  │
           └────────────────────────┘

Key Security Benefits:
✅ Unauthorized users can't determine if task 999 exists
✅ Same error for "not found" and "no permission"
✅ Prevents resource enumeration attacks
✅ Consistent security behavior across all endpoints
```

## Dependency Flow

The layers have strict one-way dependencies:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Dependency Direction                          │
└─────────────────────────────────────────────────────────────────┘

Handlers ───────────> Services ───────────> Models
(routes)                                    (data)
   │                     │                     │
   │                     │                     │
   ├─ Import services   ├─ Import models     │
   ├─ Import models     ├─ Import user       │
   ├─ Import user       ├─ Import events     │
   └─ Import echo       └─ Import xorm       └─ Import xorm


⚠️ FORBIDDEN DEPENDENCIES:
❌ Models → Services (creates import cycle)
❌ Services → Handlers (tight coupling)
❌ Models → Handlers (breaks architecture)

✅ ALLOWED BACKWARDS COMMUNICATION:
   Models ←(function variables)← Services
   
   Example:
   // pkg/models/task.go
   var TaskUpdateFunc func(s, task, u) error
   
   // pkg/services/task.go
   func init() {
       models.TaskUpdateFunc = NewTaskService().Update
   }
```

## Transaction Boundaries

Services control transaction boundaries to ensure data consistency:

```
┌─────────────────────────────────────────────────────────────────┐
│                  Transaction Management                          │
└─────────────────────────────────────────────────────────────────┘

Handler receives database session from middleware:
  handler.WithDBAndUser() → provides *xorm.Session

Service uses the session for atomic operations:

  func (ts *TaskService) BulkUpdate(s *xorm.Session, ...) error {
      // Start transaction
      err := s.Begin()
      if err != nil {
          return err
      }
      defer s.Rollback() // Auto-rollback if function exits with error
      
      // Perform multiple operations atomically
      for _, taskID := range taskIDs {
          task, err := getTaskByID(s, taskID)
          if err != nil {
              return err // Will rollback
          }
          
          err = updateTask(s, task)
          if err != nil {
              return err // Will rollback
          }
      }
      
      // All operations succeeded - commit
      return s.Commit()
  }

Benefits:
✅ All-or-nothing updates
✅ Consistent database state
✅ Atomic permission checks across multiple resources
✅ Prevents partial updates on permission failures
```

## Event Dispatching Flow

Services emit events for cross-cutting concerns:

```
┌─────────────────────────────────────────────────────────────────┐
│                      Event Flow                                  │
└─────────────────────────────────────────────────────────────────┘

Service Operation → Dispatch Event → Event Handlers
                                          │
                    ┌─────────────────────┼─────────────────────┐
                    │                     │                     │
                    ▼                     ▼                     ▼
          ┌──────────────────┐  ┌─────────────────┐  ┌────────────────┐
          │ Notification     │  │  Webhook        │  │  Analytics     │
          │ Service          │  │  Dispatcher     │  │  Tracker       │
          └──────────────────┘  └─────────────────┘  └────────────────┘

Example:

  // pkg/services/task.go
  func (ts *TaskService) Create(s, task, u) (*Task, error) {
      // ... create task in database ...
      
      // Dispatch event
      err = events.Dispatch(&TaskCreatedEvent{
          Task: task,
          User: u,
      })
      
      return task, nil
  }

Event handlers run asynchronously:
  - Send email notifications
  - Trigger webhooks
  - Update analytics
  - Sync with external services
```

## Comparison: Old vs New Architecture

### Old Architecture (Model-Heavy)

```
┌────────────────────────────────────────────────────────┐
│                  OLD PATTERN                            │
└────────────────────────────────────────────────────────┘

Handler (Web)
    │
    ▼
Model (Business Logic + Data + Permissions)
    ├─ Create(s, u)  ← 50 lines of business logic
    ├─ Update(s, u)  ← 40 lines of business logic
    ├─ Delete(s, u)  ← 30 lines of business logic
    ├─ ReadAll(s, u) ← 100+ lines of business logic
    └─ TableName()   ← Database mapping

Problems:
❌ Business logic mixed with data definitions
❌ Hard to test (need database for everything)
❌ Permission logic scattered across models
❌ Difficult to reuse logic (tied to web layer)
❌ Tight coupling between layers
```

### New Architecture (Service-Based)

```
┌────────────────────────────────────────────────────────┐
│                  NEW PATTERN                            │
└────────────────────────────────────────────────────────┘

Handler (Thin)
    │ ← Parse request, call service, format response
    ▼
Service (Business Logic)
    │ ← ALL business logic, permissions, validation
    ▼
Model (Data Only)
    ├─ Field definitions
    ├─ TableName() ← Database mapping
    └─ Deprecated facades → delegate to services

Benefits:
✅ Clean separation of concerns
✅ Easy to test each layer independently
✅ Centralized permission logic
✅ Reusable business logic
✅ Loose coupling between layers
✅ Better security (permission-before-existence)
```

## Testing Strategy by Layer

Each layer has different testing requirements:

```
┌─────────────────────────────────────────────────────────────────┐
│                     Testing Pyramid                              │
└─────────────────────────────────────────────────────────────────┘

                        ▲
                       ╱ ╲
                      ╱   ╲      E2E Tests
                     ╱     ╲     (Minimal - Full Stack)
                    ╱───────╲    
                   ╱         ╲
                  ╱           ╲  Integration Tests
                 ╱             ╲ (Service Layer Tests)
                ╱───────────────╲
               ╱                 ╲
              ╱                   ╲ Unit Tests
             ╱                     ╲ (Model Tests)
            ╱───────────────────────╲
           
Layer          │ Tests Location            │ Test Focus
──────────────┼──────────────────────────┼───────────────────────
Handlers      │ pkg/routes/api/v1/*_test │ HTTP integration
              │                           │ - Request parsing
              │                           │ - Response formatting
              │                           │ - Header handling
──────────────┼──────────────────────────┼───────────────────────
Services      │ pkg/services/*_test.go   │ Business logic (MAIN)
              │                           │ - All CRUD operations
              │                           │ - Permission checks
              │                           │ - Business rules
              │                           │ - Edge cases
              │                           │ - Error handling
              │ Target: ≥90% coverage     │
──────────────┼──────────────────────────┼───────────────────────
Models        │ pkg/models/*_test.go     │ Data structure only
              │                           │ - TableName()
              │                           │ - Field validation
              │                           │ - NO CRUD tests
              │                           │ - NO business logic

Test Setup Requirements:
- Service tests: Call testutil.Init() for dependency injection
- Model tests: DO NOT call testutil.Init() (avoid circular dependencies)
- Route tests: Use testutil.Init() + HTTP test framework
```

## Real-World Example Files

For hands-on examples, examine these files:

### Handlers (Thin Layer)
- `pkg/routes/api/v1/task.go` - Task endpoints
- `pkg/routes/api/v1/label.go` - Label endpoints
- `pkg/routes/api/v1/project.go` - Project endpoints

### Services (Business Logic)
- `pkg/services/task.go` - Task management (high complexity)
- `pkg/services/label.go` - Label management (medium complexity)
- `pkg/services/api_token.go` - API token management (low complexity)

### Models (Data Only)
- `pkg/models/task.go` - Task struct & deprecated facades
- `pkg/models/label.go` - Label struct & table mapping
- `pkg/models/project.go` - Project struct & relationships

### Tests
- `pkg/services/task_test.go` - Comprehensive service tests
- `pkg/services/label_test.go` - Service integration tests
- `pkg/models/label_test.go` - Minimal model tests

## Further Reading

- **REFACTORING_GUIDE.md** - Complete refactoring patterns and workflow
- **API_ROUTE_REFACTORING.md** - Declarative routing pattern guide
- **Section 6 of REFACTORING_GUIDE.md** - Security enhancements
- **Section 7 of REFACTORING_GUIDE.md** - Step-by-step migration guide
