# **Vikunja Service Layer Refactoring Guide**

This document is the **source of truth** for the Vikunja service layer refactor. It documents the core architectural patterns and the mandatory workflow that all agents must follow.

## **1\. Critical Agent Instructions & Workflow**

This section contains the most important, non-negotiable rules. You must understand and follow these at all times.

### **The Golden Rule: Move Logic, Don't Expose It**

This is the most important rule of the entire refactor. When a service needs logic that is currently inside a private models function, the solution is **always** to **MOVE** that logic into the service.

You must **NEVER** solve this by making the model function public and calling it from the service.

* **Correct ✅:** Copying the permission-checking code from a model method into a new service method.  
* **Incorrect ❌:** Making a model's CanRead() or CanWrite() method public and calling it from a service.

### **Plan Management & Progress Reporting**

Once your plan is approved, you must use it as your active checklist. The following process is mandatory:

1. **State Your Step:** At the beginning of each action, you must state which step of the approved plan you are working on.  
2. **Update on Change:** If you encounter a problem, you must first state the problem and then **use your tool to update the plan** before proceeding.  
3. **Final Report:** Before submitting your work, you must provide a final summary of the plan, marking each step as complete (✅).

### **Technical Debt Management**

Any shortcut or deviation from the service layer architecture must be immediately documented as technical debt:

1. **Document Immediately:** Create follow-up tasks with clear descriptions of the shortcuts taken
2. **Label Technical Debt:** Use "Technical Debt" prefix in task titles for easy identification  
3. **Block Next Phase:** Technical debt tasks must be completed before moving to subsequent phases
4. **Example:** If you call `models.AddMoreInfoToTasks` from a service instead of implementing proper service methods, create tasks like "Technical Debt: Implement Service Layer Expansion Methods"

### **Running Tests: The Pre-Flight Checklist**

You must follow these steps **every time** you need to run tests to avoid common environment failures.

1. **Enter the Dev Shell:** Ensure you are in the correct environment by running:  
   devenv shell

2. **Set the Root Path:** Before running any test command, you **must** set the environment variable. This is critical for loading test fixtures.  
   export VIKUNJA\_SERVICE\_ROOTPATH=$(pwd)

3. **Build Frontend (for Web Tests ONLY):** If you are about to run the full end-to-end web tests (mage test:web), you must build the frontend assets first.  
   cd frontend && pnpm install && pnpm run build && cd ..

4. **Run the Correct Command:**  
   * For backend-only tests: mage test:feature  
   * For end-to-end web tests: mage test:web  
   * For a specific package: go test ./pkg/services/...

## **2\. The Core Architecture ("Chef, Waiter, Pantry")**

Our architecture separates concerns into three distinct layers:

* **🧑‍🍳 Services (pkg/services) \- The "Chef":** Contains **ALL** business logic. Decoupled from the web layer, uses \*user.User.  
* **🧑‍💼 Handlers (pkg/routes) \- The "Waiter":** A thin "glue" layer. Parses requests, calls services, returns responses. Uses **handler wrappers**.  
* **🏪 Models (pkg/models) \- The "Pantry":** A "dumb" data layer for basic database access (CRUD) only. Contains **no** business logic.

## **3\. Key Patterns & Implementation**

### A. Declarative Routing & Handler Wrappers

To keep handlers clean and permissions explicit, we use a declarative routing pattern combined with a handler wrapper. This is the **required pattern** for all new handlers.

This pattern uses the `APIRoute` struct and a `routes.Register` helper function to define all routes and their permission scopes in a single, clear, and maintainable way.

**For a complete explanation and implementation guide for this pattern, you must refer to the `API_ROUTE_REFACTORING.md` file.**

**Example (`pkg/routes/api/v1/label.go`):**

```go
var LabelRoutes = []routes.APIRoute{
    {Method: "GET",  Path: "/labels", Handler: handler.WithDBAndUser(getAllLabelsLogic, false), PermissionScope: "read_all"},
    {Method: "POST", Path: "/labels", Handler: handler.WithDBAndUser(createLabelLogic, true),  PermissionScope: "create"},
    // ... all other label routes are defined in this slice
}

func RegisterLabels(a *echo.Group) {
    routes.Register(a, LabelRoutes)
}
```

### **B. Dependency Inversion (For Backward Compatibility)**

To break the models \-\> services import cycle when deprecating model methods, we use this pattern.

**1\. In the Model (pkg/models/tasks.go):**

var TaskUpdateFunc func(s \*xorm.Session, t \*Task, u \*user.User) error

// @Deprecated: Use services.TaskService.Update instead.  
func (t \*Task) Update(s \*xorm.Session, u \*user.User) error { /\* calls TaskUpdateFunc \*/ }

**2\. In the Service (pkg/services/task.go):**

func init() {  
    models.TaskUpdateFunc \= NewTaskService(nil).Update  
}

## **4\. The Testing Philosophy**

Our testing strategy is layered to match the architecture.

* **Model Tests (pkg/models/):** Simple **unit tests** only. They must remain fully decoupled from the service layer.  
* **Service & Integration Tests (pkg/services/, pkg/webtests/):** This is the home for our complex **integration tests**. They verify the complete business logic and end-to-end functionality.

### **Test Environment Setup**

#### **For Service & Integration Tests (Most Packages)**

These test suites need the full application to be "wired up." Their TestMain function **must** call our "Master Switch" to initialize all service dependencies.

// In a main\_test.go file for a package like \`services\` or \`migration\`  
import "code.vikunja.io/api/pkg/testutil"

func TestMain(m \*testing.M) {  
    // ... other setup ...  
    testutil.Init() // Initialize service dependency injection  
    // ... rest of setup ...  
}

#### **For Model Tests (The Exception)**

The pkg/models test suite **must not** call testutil.Init(), as this would create a circular dependency. Instead, if a model test needs to verify a function that calls a dependency-injected function variable, we provide a **simple mock** in the test setup.

// In pkg/models/main\_test.go  
func TestMain(m \*testing.M) {  
    // ... other setup ...

    // Set up a mock for the TaskCreateFunc for model tests,  
    // as they should not depend on the services package.  
    TaskCreateFunc \= func(s \*xorm.Session, t \*Task, u \*user.User) error {  
        // This is a simple mock. It does not contain real logic.  
        t.ID \= 999 // Give it a fake ID to signify success  
        return nil  
    }

    // ... rest of TestMain ...  
    os.Exit(m.Run())  
}  

## **5\. Testing Strategy for Refactored Components**

After completing the service layer refactor (Phase 2.1-2.2), the testing strategy has been updated to eliminate duplication and focus on testing actual system behavior.

### **DO NOT Test Deprecated Model Methods**

- Model CRUD methods (Create, Update, Delete, ReadAll) are deprecated facades
- They delegate to service layer with zero business logic
- Testing them validates the mock, not the system
- **These tests have been removed** from model test files

### **Test at Service Layer Instead**

- **Business logic tests** → `pkg/services/*_test.go`
  - Test actual service implementations
  - Comprehensive coverage of all business logic paths
  - Integration with database using real transactions
  
- **Integration tests** → Service tests with `testutil.Init()`
  - Full application context with dependency injection
  - Real service interactions, not mocks
  - End-to-end validation of business flows

- **Route tests** → `pkg/routes/api/v1/*_test.go` (if needed)
  - HTTP-level integration tests
  - Request/response validation
  - Authentication and permission flows

### **Model Tests Should Only Cover**

Model tests in `pkg/models/*_test.go` should be minimal and focused on:

- **TableName() function** - Database table name mapping
- **Struct field validation** - Pure data structure behavior (not database operations)
- **Helper functions** - Temporarily kept until T-PERMISSIONS task
  - Examples: `GetAPITokenByID`, `GetTokenFromTokenString`, `getLabelByIDSimple`
  - These will be moved to service layer in T-PERMISSIONS
- **Permission methods** - Temporarily kept until T-PERMISSIONS task
  - Examples: `CanRead`, `CanUpdate`, `CanDelete`
  - These will be moved to service layer in T-PERMISSIONS

### **What Has Been Removed**

The following have been removed from model tests to eliminate technical debt:

- ✅ Mock services (`mockFavoriteService`, `mockLabelService`, `mockAPITokenService`, `mockReactionsService`, `mockProjectTeamService`, `mockProjectUserService`)
- ✅ CRUD method tests (Create, Update, Delete, ReadAll tests for deprecated methods)
- ✅ Business logic tests that duplicate service layer tests

### **Benefits of This Approach**

1. **No Mock Maintenance** - Service tests use real implementations
2. **Faster Test Execution** - Fewer redundant tests
3. **Better Coverage** - Tests validate actual system behavior, not scaffolding
4. **Clear Separation** - Service tests for logic, model tests for structure
5. **Easier Refactoring** - Change service implementation without breaking model tests

### **Example: Before and After**

**Before (Testing Mocks):**
```go
// pkg/models/label_test.go
func TestLabel_Create(t *testing.T) {
    // This tests mockLabelService.Create, not the real system
    label := &Label{Title: "Test"}
    err := label.Create(s, user)  // Calls mock, not real service
    assert.NoError(t, err)
}
```

**After (Testing Real Service):**
```go
// pkg/services/label_test.go
func TestLabelService_Create(t *testing.T) {
    // This tests actual LabelService implementation
    service := NewLabelService(db)
    label := &models.Label{Title: "Test"}
    err := service.Create(s, label, user)  // Tests real business logic
    assert.NoError(t, err)
    // Can verify database state, events, etc.
}
```

### **Migration Status**

- ✅ **Phase 2.1-2.2 Complete** - All mock services removed, CRUD tests deleted
- ⚠️ **T-PERMISSIONS Pending** - Helper and permission methods still in models
- 📋 **Future** - After T-PERMISSIONS, models will be pure data structures with zero database operations

## **6. Security Enhancements in Service Layer**

The service layer refactor introduced several important security improvements over the original model-based architecture. Understanding these improvements helps maintain consistent security patterns across the codebase.

### **Permission Checks Before Existence Checks**

**Security Issue in Model Layer:**
The original model layer often checked if a resource exists before checking if the user has permission to access it. This creates an **information disclosure vulnerability** where unauthorized users can determine if a resource exists.

**Example - Model Layer (Vulnerable):**
```go
// pkg/models/task.go (OLD PATTERN - DO NOT USE)
func (t *Task) CanRead(s *xorm.Session, u *user.User) bool {
    // Check if task exists
    exists, _ := s.Where("id = ?", t.ID).Exist(&Task{})
    if !exists {
        return false  // Reveals that task doesn't exist
    }
    
    // Then check permissions
    return checkProjectPermission(s, t.ProjectID, u, PermissionRead)
}
```

**Problem:** An attacker can probe task IDs and learn which tasks exist, even if they don't have access.

**Service Layer (Secure):**
```go
// pkg/services/task.go (NEW PATTERN - SECURE)
func (s *TaskService) Get(sess *xorm.Session, taskID int64, u *user.User) (*models.Task, error) {
    // Get task first (without revealing if it exists)
    task, err := getTaskByID(sess, taskID)
    if err != nil {
        // Don't reveal if it's "not found" vs "forbidden"
        return nil, ErrGenericForbidden{}
    }
    
    // Check permission BEFORE revealing existence
    hasAccess, err := s.checkTaskPermission(sess, task, u, PermissionRead)
    if err != nil || !hasAccess {
        return nil, ErrGenericForbidden{}  // Same error for "doesn't exist" and "no permission"
    }
    
    return task, nil
}
```

**Key Improvement:** Unauthorized users receive `403 Forbidden` whether the task exists or not, preventing information leakage.

### **Consistent Error Messages for Security**

**Security Pattern:** Always return `ErrGenericForbidden` for both "resource doesn't exist" and "user lacks permission" scenarios. This prevents attackers from enumerating valid resource IDs.

**Example - Consistent Error Handling:**
```go
// Service layer security pattern
func (s *TaskService) Update(sess *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
    // Fetch task
    existingTask, err := getTaskByID(sess, task.ID)
    if err != nil {
        return nil, ErrGenericForbidden{}  // Don't reveal "not found"
    }
    
    // Check permission
    canUpdate, err := s.checkTaskPermission(sess, existingTask, u, PermissionUpdate)
    if err != nil || !canUpdate {
        return nil, ErrGenericForbidden{}  // Same error for all unauthorized cases
    }
    
    // Proceed with update...
}
```

**Migration Note:** When refactoring model code to services, replace existence-revealing errors like `ErrTaskDoesNotExist` with `ErrGenericForbidden` when permission checks are involved.

### **Link Share Permission Handling**

**Security Enhancement:** The service layer properly validates link share tokens and their expiration before granting access.

**Service Layer Pattern:**
```go
// pkg/services/task.go
func (s *TaskService) getTaskWithLinkShare(sess *xorm.Session, taskID int64, linkShareToken string) (*models.Task, error) {
    // Validate link share exists and is not expired
    linkShare, err := s.LinkShareService.GetByToken(sess, linkShareToken)
    if err != nil {
        return nil, ErrGenericForbidden{}  // Invalid or expired token
    }
    
    // Check that link share grants access to this specific task
    if !linkShare.GrantsAccessTo(taskID) {
        return nil, ErrGenericForbidden{}  // Token doesn't grant access to this resource
    }
    
    // Fetch and return task
    return getTaskByID(sess, taskID)
}
```

**Key Improvements:**
1. Token validation happens before database queries
2. Expired tokens are rejected immediately
3. Scope checking ensures tokens only grant intended access
4. Consistent error responses prevent token enumeration

### **Transaction Boundary Security**

**Service Layer Advantage:** Services control transaction boundaries, ensuring atomic security operations.

**Example - Secure Bulk Operations:**
```go
// pkg/services/task.go
func (s *TaskService) BulkUpdate(sess *xorm.Session, updates *BulkUpdateRequest, u *user.User) error {
    // Start transaction
    err := sess.Begin()
    if err != nil {
        return err
    }
    defer sess.Rollback()
    
    // Check permissions for ALL tasks BEFORE making any changes
    for _, taskID := range updates.TaskIDs {
        task, err := getTaskByID(sess, taskID)
        if err != nil {
            return ErrGenericForbidden{}
        }
        
        hasPermission, err := s.checkTaskPermission(sess, task, u, PermissionUpdate)
        if err != nil || !hasPermission {
            return ErrGenericForbidden{}  // Abort entire operation if any task unauthorized
        }
    }
    
    // All permission checks passed - now apply updates
    for _, taskID := range updates.TaskIDs {
        // Apply updates...
    }
    
    return sess.Commit()
}
```

**Security Benefit:** Permission failures roll back the entire operation - no partial updates that might leak information.

### **Best Practices Summary**

When implementing service layer methods:

1. ✅ **Check permissions FIRST** - Before revealing resource existence
2. ✅ **Use consistent errors** - `ErrGenericForbidden` for all unauthorized access
3. ✅ **Validate tokens early** - Before expensive database operations
4. ✅ **Control transactions** - Ensure atomic permission checks and updates
5. ✅ **Avoid information leakage** - Don't distinguish between "not found" and "forbidden"
6. ✅ **Document security decisions** - Explain why security checks are ordered as they are

### **Testing Security Improvements**

Service layer tests should verify security behavior:

```go
// pkg/services/task_test.go
func TestTaskService_Get_Security(t *testing.T) {
    t.Run("forbidden error for non-existent task", func(t *testing.T) {
        // User tries to access non-existent task
        task, err := taskService.Get(sess, 99999, user)
        
        assert.Nil(t, task)
        assert.Error(t, err)
        assert.IsType(t, ErrGenericForbidden{}, err)  // Should be forbidden, not "not found"
    })
    
    t.Run("forbidden error for unauthorized access", func(t *testing.T) {
        // User tries to access task in project they don't have access to
        task, err := taskService.Get(sess, existingTaskID, unauthorizedUser)
        
        assert.Nil(t, task)
        assert.Error(t, err)
        assert.IsType(t, ErrGenericForbidden{}, err)  // Same error as non-existent
    })
}
```

**See Also:**
- Phase 2 implementation examples in `pkg/services/task.go`, `pkg/services/project.go`
- Security test examples in `pkg/services/task_business_logic_test.go`
- Comprehensive security validation in `TestTaskService_Assignee_WithoutProjectAccess`

## **7. Step-by-Step Service Layer Migration Guide**

This section provides a comprehensive, step-by-step guide for migrating a feature from the model layer to the service layer. Follow this process to ensure architectural consistency and avoid common pitfalls.

### **7.1 Pre-Migration Assessment**

Before starting the migration, assess the feature:

1. **Identify all related files:**
   ```bash
   # Find model files
   find pkg/models -name "*feature_name*.go"
   
   # Check for existing service files
   find pkg/services -name "*feature_name*.go"
   
   # Find route handlers
   grep -r "feature" pkg/routes/
   ```

2. **Map dependencies:**
   - What other models does this feature depend on?
   - What features depend on this feature?
   - Are there circular dependencies?

3. **Document current behavior:**
   - Run existing tests and document their coverage
   - Note any quirks or edge cases in current implementation

### **7.2 Service Creation**

**Step 1: Create the service file**

```go
// pkg/services/feature.go
package services

import (
    "code.vikunja.io/api/pkg/models"
    "code.vikunja.io/api/pkg/user"
    "xorm.io/xorm"
)

// FeatureService handles all business logic for features
type FeatureService struct {
    DB *xorm.Engine
    // Add dependent services as needed
}

// NewFeatureService creates a new FeatureService instance
func NewFeatureService(db *xorm.Engine) *FeatureService {
    return &FeatureService{
        DB: db,
    }
}
```

**Step 2: Move CRUD operations**

For each model CRUD method (Create, ReadOne, ReadAll, Update, Delete):

1. **Copy the business logic** from model to service
2. **Add permission checks** at the service level
3. **Keep model method as deprecated facade** (for backward compatibility)

**Example - Moving Create method:**

```go
// pkg/services/feature.go
func (fs *FeatureService) Create(s *xorm.Session, feature *models.Feature, u *user.User) (*models.Feature, error) {
    // 1. Validate input
    if feature.Title == "" {
        return nil, ErrFeatureTitleCannotBeEmpty{}
    }
    
    // 2. Check permissions (permission-before-existence pattern)
    canCreate, err := fs.checkCreatePermission(s, feature, u)
    if err != nil || !canCreate {
        return nil, ErrGenericForbidden{}
    }
    
    // 3. Set metadata
    feature.Created = time.Now()
    feature.CreatedByID = u.ID
    
    // 4. Perform database operation
    _, err = s.Insert(feature)
    if err != nil {
        return nil, err
    }
    
    // 5. Emit events if needed
    err = events.Dispatch(&FeatureCreatedEvent{Feature: feature})
    if err != nil {
        return nil, err
    }
    
    return feature, nil
}
```

**Step 3: Deprecate model methods**

```go
// pkg/models/feature.go
// @Deprecated: Use services.FeatureService.Create instead
// This method is maintained for backward compatibility only
func (f *Feature) Create(s *xorm.Session, u *user.User) error {
    if FeatureCreateFunc != nil {
        _, err := FeatureCreateFunc(s, f, u)
        return err
    }
    return errors.New("FeatureCreateFunc not initialized")
}
```

### **7.3 Complex Business Logic Migration**

**For methods with complex logic (filters, sorting, pagination):**

1. **Extract query building logic** into service
2. **Move permission filtering** to service layer
3. **Keep query execution** in service

**Example - Moving ReadAll with filters:**

```go
// pkg/services/feature.go
func (fs *FeatureService) GetAll(s *xorm.Session, opts *models.FeatureSearchOptions, u *user.User) ([]*models.Feature, int, int64, error) {
    // 1. Build base query
    query := s.Where("1 = 1")
    
    // 2. Apply permission filters
    projectIDs, err := fs.getAccessibleProjectIDs(s, u)
    if err != nil {
        return nil, 0, 0, err
    }
    query = query.In("project_id", projectIDs)
    
    // 3. Apply search filters
    if opts.Search != "" {
        query = query.Where("title LIKE ?", "%"+opts.Search+"%")
    }
    
    // 4. Apply sorting
    if opts.SortBy != "" {
        query = query.OrderBy(opts.SortBy + " " + opts.SortOrder)
    }
    
    // 5. Get total count before pagination
    totalItems, err := query.Count(&models.Feature{})
    if err != nil {
        return nil, 0, 0, err
    }
    
    // 6. Apply pagination
    query = query.Limit(opts.PerPage, (opts.Page-1)*opts.PerPage)
    
    // 7. Execute query
    var features []*models.Feature
    err = query.Find(&features)
    if err != nil {
        return nil, 0, 0, err
    }
    
    return features, len(features), totalItems, nil
}
```

### **7.4 Dependency Injection Setup**

**Step 1: Define function variables in models**

```go
// pkg/models/feature.go
var (
    FeatureCreateFunc func(s *xorm.Session, f *Feature, u *user.User) (*Feature, error)
    FeatureUpdateFunc func(s *xorm.Session, f *Feature, u *user.User) (*Feature, error)
    FeatureDeleteFunc func(s *xorm.Session, f *Feature, u *user.User) error
)
```

**Step 2: Wire services in service init function**

```go
// pkg/services/feature.go
func InitFeatureService() {
    models.FeatureCreateFunc = func(s *xorm.Session, f *models.Feature, u *user.User) (*models.Feature, error) {
        return NewFeatureService(s.Engine()).Create(s, f, u)
    }
    
    models.FeatureUpdateFunc = func(s *xorm.Session, f *models.Feature, u *user.User) (*models.Feature, error) {
        return NewFeatureService(s.Engine()).Update(s, f, u)
    }
    
    models.FeatureDeleteFunc = func(s *xorm.Session, f *models.Feature, u *user.User) error {
        return NewFeatureService(s.Engine()).Delete(s, f.ID, u)
    }
}
```

**Step 3: Call init in testutil**

```go
// pkg/testutil/testutil.go
func Init() {
    // ... other service inits ...
    services.InitFeatureService()
}
```

### **7.5 Route Migration (Declarative Pattern)**

**Step 1: Create route file**

```go
// pkg/routes/api/v1/feature.go
package v1

import (
    "code.vikunja.io/api/pkg/models"
    "code.vikunja.io/api/pkg/routes"
    "code.vikunja.io/api/pkg/routes/api/handler"
    "code.vikunja.io/api/pkg/services"
    "code.vikunja.io/api/pkg/user"
    "github.com/labstack/echo/v4"
    "xorm.io/xorm"
)

var FeatureRoutes = []routes.APIRoute{
    {Method: "GET", Path: "/features", Handler: handler.WithDBAndUser(getAllFeaturesLogic, false), PermissionScope: "read_all"},
    {Method: "POST", Path: "/features", Handler: handler.WithDBAndUser(createFeatureLogic, true), PermissionScope: "create"},
    {Method: "GET", Path: "/features/:feature", Handler: handler.WithDBAndUser(getFeatureLogic, false), PermissionScope: "read"},
    {Method: "PUT", Path: "/features/:feature", Handler: handler.WithDBAndUser(updateFeatureLogic, true), PermissionScope: "update"},
    {Method: "DELETE", Path: "/features/:feature", Handler: handler.WithDBAndUser(deleteFeatureLogic, true), PermissionScope: "delete"},
}

func RegisterFeatures(a *echo.Group) {
    routes.Register(a, FeatureRoutes)
}
```

**Step 2: Implement handler logic functions**

```go
// pkg/routes/api/v1/feature.go (continued)
func getAllFeaturesLogic(s *xorm.Session, u *user.User, c echo.Context) (interface{}, error) {
    // 1. Parse query parameters
    opts := &models.FeatureSearchOptions{}
    if err := c.Bind(opts); err != nil {
        return nil, err
    }
    
    // 2. Call service
    featureService := services.NewFeatureService(s.Engine())
    features, count, total, err := featureService.GetAll(s, opts, u)
    if err != nil {
        return nil, err
    }
    
    // 3. Set pagination headers
    c.Response().Header().Set("x-pagination-total-items", strconv.FormatInt(total, 10))
    c.Response().Header().Set("x-pagination-result-count", strconv.Itoa(count))
    
    return features, nil
}

func createFeatureLogic(s *xorm.Session, u *user.User, c echo.Context) (interface{}, error) {
    // 1. Parse request body
    feature := &models.Feature{}
    if err := c.Bind(feature); err != nil {
        return nil, err
    }
    
    // 2. Call service
    featureService := services.NewFeatureService(s.Engine())
    created, err := featureService.Create(s, feature, u)
    if err != nil {
        return nil, err
    }
    
    return created, nil
}
```

**Step 3: Register routes**

```go
// pkg/routes/routes.go
import v1 "code.vikunja.io/api/pkg/routes/api/v1"

func RegisterRoutes(a *echo.Group) {
    // ... other routes ...
    v1.RegisterFeatures(a)
}
```

### **7.6 Test Migration**

**Step 1: Create service tests**

```go
// pkg/services/feature_test.go
package services

import (
    "testing"
    
    "code.vikunja.io/api/pkg/db"
    "code.vikunja.io/api/pkg/models"
    "code.vikunja.io/api/pkg/user"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFeatureService_Create(t *testing.T) {
    t.Run("create feature successfully", func(t *testing.T) {
        // Setup
        db.LoadFixtures()
        s := db.NewSession()
        defer s.Close()
        
        u := &user.User{ID: 1}
        feature := &models.Feature{
            Title: "Test Feature",
            ProjectID: 1,
        }
        
        // Execute
        featureService := NewFeatureService(db.GetEngine())
        created, err := featureService.Create(s, feature, u)
        
        // Verify
        require.NoError(t, err)
        assert.NotZero(t, created.ID)
        assert.Equal(t, "Test Feature", created.Title)
        assert.Equal(t, u.ID, created.CreatedByID)
    })
    
    t.Run("forbidden when user lacks permission", func(t *testing.T) {
        // Setup
        db.LoadFixtures()
        s := db.NewSession()
        defer s.Close()
        
        u := &user.User{ID: 999} // User without access
        feature := &models.Feature{
            Title: "Test Feature",
            ProjectID: 1,
        }
        
        // Execute
        featureService := NewFeatureService(db.GetEngine())
        created, err := featureService.Create(s, feature, u)
        
        // Verify
        assert.Nil(t, created)
        assert.Error(t, err)
        assert.IsType(t, ErrGenericForbidden{}, err)
    })
}
```

**Step 2: Remove redundant model tests**

```go
// pkg/models/feature_test.go
// DELETE: Tests for CRUD methods (now tested in service layer)
// KEEP: Tests for TableName(), field validation, helper functions
```

### **7.7 Common Pitfalls and Solutions**

**Pitfall 1: Exposing model helper functions**
- ❌ **Wrong:** Making `models.getFeatureByID()` public
- ✅ **Right:** Copy the function to service as private method

**Pitfall 2: Testing deprecated model methods**
- ❌ **Wrong:** Writing comprehensive tests for `feature.Create()`
- ✅ **Right:** Test `FeatureService.Create()`, minimal test for model delegation

**Pitfall 3: Forgetting dependency injection**
- ❌ **Wrong:** Model methods calling services directly (import cycle)
- ✅ **Right:** Use function variables and wire in `InitFeatureService()`

**Pitfall 4: Inconsistent error handling**
- ❌ **Wrong:** Returning `ErrFeatureNotFound` when user lacks permission
- ✅ **Right:** Return `ErrGenericForbidden` for all unauthorized access

**Pitfall 5: Partial migration**
- ❌ **Wrong:** Migrating Create but leaving Update in model
- ✅ **Right:** Migrate all CRUD operations together for consistency

### **7.8 Migration Checklist**

Use this checklist to verify complete migration:

**Service Layer:**
- [ ] Service file created: `pkg/services/feature.go`
- [ ] Service struct defined with DB and dependencies
- [ ] `NewFeatureService()` constructor implemented
- [ ] All CRUD methods migrated to service
- [ ] Business logic moved from models to service
- [ ] Permission checks implemented in service
- [ ] Dependency injection setup in `InitFeatureService()`

**Model Layer:**
- [ ] CRUD methods marked as `@Deprecated`
- [ ] Model methods delegate to function variables
- [ ] Function variables declared in model file
- [ ] Helper functions remain in model (until T-PERMISSIONS)
- [ ] Permission methods remain in model (until T-PERMISSIONS)

**Routes Layer:**
- [ ] Route file created: `pkg/routes/api/v1/feature.go`
- [ ] `FeatureRoutes` slice defined with all endpoints
- [ ] Handler logic functions implemented
- [ ] `RegisterFeatures()` function implemented
- [ ] Routes registered in `pkg/routes/routes.go`
- [ ] Legacy WebHandler removed from `routes.go`

**Testing:**
- [ ] Service tests created: `pkg/services/feature_test.go`
- [ ] All business logic paths covered in service tests
- [ ] Security tests added (permission checks)
- [ ] Edge case tests implemented
- [ ] Model CRUD tests removed
- [ ] Tests pass: `VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test ./pkg/services`

**Integration:**
- [ ] `testutil.Init()` calls `InitFeatureService()`
- [ ] Application compiles: `go build`
- [ ] No import cycles: `go list -f '{{.ImportPath}} {{.Imports}}' ./... | grep cycle`
- [ ] All tests pass: `VIKUNJA_SERVICE_ROOTPATH=$(pwd) mage test:feature`

**Documentation:**
- [ ] Swagger annotations updated in route handlers
- [ ] Deprecation comments added to model methods
- [ ] Any architectural decisions documented

### **7.9 Example Migrations to Study**

For real-world examples of successful migrations, review these completed features:

1. **Simple Feature:** Labels (`pkg/services/label.go`, `pkg/routes/api/v1/label.go`)
2. **Medium Complexity:** API Tokens (`pkg/services/api_token.go`, `pkg/routes/api/v1/api_token.go`)
3. **High Complexity:** Tasks (`pkg/services/task.go`, `pkg/routes/api/v1/task.go`)
4. **With Dependencies:** Project Teams (`pkg/services/project_team.go`)

Each migration demonstrates:
- Clean separation of concerns
- Proper permission handling
- Comprehensive test coverage
- Declarative routing pattern

