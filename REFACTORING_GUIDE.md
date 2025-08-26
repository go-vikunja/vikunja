# Vikunja Service Layer Refactoring Guide

This document serves as the **source of truth** for all refactoring work in the Vikunja project. It documents the core architectural patterns, dependency management strategies, and testing philosophies established during the service layer refactor.

## Table of Contents

1. [The Core Architecture ("Chef, Waiter, Pantry")](#the-core-architecture-chef-waiter-pantry)
2. [The Dependency Inversion Pattern](#the-dependency-inversion-pattern)
3. [The Testing Philosophy](#the-testing-philosophy)
4. [Development Environment Setup](#development-environment-setup)
5. [Common Patterns and Examples](#common-patterns-and-examples)

---

## The Core Architecture ("Chef, Waiter, Pantry")

The Vikunja refactor follows a three-layer architecture pattern that cleanly separates concerns:

### 🧑‍🍳 Services (`pkg/services`) - The "Chef"

**Role**: Contains ALL business logic, validation, permission checks, and complex operations.

**Responsibilities**:
- Business logic implementation
- Permission and access control
- Data validation and transformation
- Complex database operations
- Event dispatching
- Integration between different models

**Example**:
```go
// pkg/services/project.go
func (p *Project) Update(s *xorm.Session, project *models.Project, u *user.User) (*models.Project, error) {
    // Permission check
    can, err := project.CanUpdate(s, u)
    if err != nil {
        return nil, err
    }
    if !can {
        return nil, &models.ErrGenericForbidden{}
    }

    // Business logic validation
    err = p.validate(s, project)
    if err != nil {
        return nil, err
    }

    // Complex business logic: cascade archive to children
    if project.IsArchived {
        err = SetArchiveStateForProjectDescendants(s, project.ID, project.IsArchived)
        if err != nil {
            return nil, err
        }
    }

    // Database update
    _, err = s.ID(project.ID).Cols("title", "is_archived", "identifier").Update(project)
    return project, err
}
```

### 🧑‍💼 Handlers (`pkg/routes`) - The "Waiter"

**Role**: Thin glue layer between HTTP API and services. Should contain minimal logic.

**Responsibilities**:
- HTTP request/response handling
- Parameter extraction and basic validation
- Calling appropriate service methods
- Error handling and HTTP status code mapping
- Response formatting

**Anti-patterns to avoid**:
- Business logic in handlers
- Direct database operations
- Complex validation logic

### 🏪 Models (`pkg/models`) - The "Pantry"

**Role**: "Dumb" data layer for basic database operations only.

**Responsibilities**:
- Database schema definitions
- Basic CRUD operations
- Simple data validation (field presence, format)
- Database relationships and constraints

**What models should NOT contain**:
- Complex business logic
- Permission checks
- Multi-model operations
- Event dispatching

**Example of proper model method**:
```go
// pkg/models/project.go - Simple, delegated method
func (p *Project) Update(s *xorm.Session, a web.Auth) (err error) {
    // Delegate to service layer via dependency inversion
    if ProjectUpdateFunc != nil {
        u, err := user.GetFromAuth(a)
        if err != nil {
            return err
        }
        _, err = ProjectUpdateFunc(s, p, u)
        return err
    }
    return errors.New("ProjectUpdateFunc not initialized")
}
```

---

## The Dependency Inversion Pattern

### The Problem

When refactoring from models to services, we encounter a circular import problem:
- `models` package needs to call `services` for complex operations
- `services` package already imports `models` for data structures
- Direct import creates: `models` → `services` → `models` (circular dependency)

### The Solution: Function Variables + init()

We use **dependency inversion** with function variables to break the cycle:

#### Step 1: Declare Function Variable in Models

```go
// pkg/models/project.go
package models

// ProjectUpdateFunc is a function variable that can be set by the services package
// to provide the implementation for project updates. This breaks the circular import.
var ProjectUpdateFunc func(s *xorm.Session, project *Project, u *user.User) (*Project, error)
```

#### Step 2: Use Function Variable in Model Method

```go
// pkg/models/project.go
func (p *Project) Update(s *xorm.Session, a web.Auth) (err error) {
    if ProjectUpdateFunc != nil {
        u, err := user.GetFromAuth(a)
        if err != nil {
            return err
        }
        _, err = ProjectUpdateFunc(s, p, u)
        return err
    }
    return errors.New("ProjectUpdateFunc not initialized")
}
```

#### Step 3: Set Function Variable in Services init()

```go
// pkg/services/project.go
package services

import "code.vikunja.io/api/pkg/models"

func init() {
    // Set up dependency injection for models to use service layer functions
    models.ProjectUpdateFunc = func(s *xorm.Session, project *models.Project, u *user.User) (*models.Project, error) {
        projectService := &Project{DB: s.Engine()}
        return projectService.Update(s, project, u)
    }
}
```

### Key Benefits

1. **No circular imports**: Models never import services directly
2. **Runtime injection**: Services inject their implementation at runtime
3. **Clean separation**: Models remain thin, services contain logic
4. **Testable**: Can inject test implementations if needed

---

## The Testing Philosophy

### Model Tests: Simple Unit Tests

**Location**: `pkg/models/*_test.go`

**Purpose**: Test basic model functionality, simple validation, and database operations.

**Characteristics**:
- Fast execution
- Minimal dependencies
- Focus on data validation and basic CRUD
- Should NOT test complex business logic

**Example**:
```go
func TestProject_CreateOrUpdate_Create_Normal(t *testing.T) {
    db.LoadAndAssertFixtures(t)
    s := db.NewSession()
    project := Project{
        Title:       "test",
        Description: "Lorem Ipsum",
    }
    err := project.Create(s, usr)
    assert.NoError(t, err)
    assert.NotEqual(t, 0, project.ID)
}
```

### Service Tests: Integration Tests

**Location**: `pkg/services/*_test.go`

**Purpose**: Test complete business logic, complex operations, and integration between components.

**Characteristics**:
- Test real business scenarios
- Include permission checks
- Test complex operations (like cascading archives)
- Verify complete workflows

**Example**:
```go
func TestProject_Update_ArchiveParentArchivesChild(t *testing.T) {
    db.LoadAndAssertFixtures(t)
    s := db.NewSession()
    defer s.Close()

    p := &Project{DB: db.GetEngine()}
    actingUser := &user.User{ID: 6}

    // Load existing project and set archive flag
    existingProject, err := models.GetProjectSimpleByID(s, 27)
    require.NoError(t, err)
    existingProject.IsArchived = true

    // Test the complete business logic
    updatedProject, err := p.Update(s, existingProject, actingUser)
    assert.NoError(t, err)
    assert.True(t, updatedProject.IsArchived)

    // Verify cascading behavior
    db.AssertExists(t, "projects", map[string]interface{}{
        "id":          27,
        "is_archived": true,
    }, false)
    db.AssertExists(t, "projects", map[string]interface{}{
        "id":          12, // Child project
        "is_archived": true,
    }, false)
}
```

### Test Setup Pattern

Each test package should have a `TestMain` function for environment setup:

#### Models Test Setup

```go
// pkg/models/main_test.go
func TestMain(m *testing.M) {
    setupTime()
    log.InitLogger()
    config.InitDefaultConfig()
    config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))
    i18n.Init()
    files.InitTests()
    user.InitTests()
    SetupTests()
    events.Fake()
    os.Exit(m.Run())
}
```

#### Services Test Setup

```go
// pkg/services/main_test.go
func TestMain(m *testing.M) {
    log.InitLogger()
    config.InitDefaultConfig()
    config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))
    i18n.Init()
    files.InitTests()
    user.InitTests()
    models.SetupTests()  // Use models setup for database
    events.Fake()
    os.Exit(m.Run())
}
```

### Error Type Consistency

When implementing validation in services, ensure error types match what the model tests expect:

```go
// Check if error checker expects pointer or value type
func IsErrProjectIdentifierIsNotUnique(err error) bool {
    _, ok := err.(ErrProjectIdentifierIsNotUnique)  // Value type
    return ok
}

// Return matching type in validation
if exists {
    return ErrProjectIdentifierIsNotUnique{Identifier: project.Identifier}  // Value, not pointer
}
```

---

## Development Environment Setup

### Running Tests

**Important**: Always use the `devenv shell` within WSL for proper Go environment:

```bash
# Start WSL
wsl

# Enter development environment
devenv shell

# Set environment variable and run tests
VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test -v ./pkg/services
VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test -v ./pkg/models
```

### Common Issues and Solutions

1. **Import cycle errors**: Use dependency inversion pattern (see above)
2. **"go: command not found"**: Ensure you're in `devenv shell`
3. **Fixture loading failures**: Set `VIKUNJA_SERVICE_ROOTPATH` environment variable
4. **Unused import errors**: Clean up imports after refactoring

---

## Common Patterns and Examples

### Deprecating Model Methods

When moving logic from models to services:

1. **Add deprecation comment**:
```go
// @Deprecated: This logic has been moved to the service layer.
// Use services.Project.Update instead of this model-level function.
func UpdateProject(s *xorm.Session, project *Project, auth web.Auth) error {
    // ... existing implementation for backward compatibility
}
```

2. **Create service method**:
```go
// pkg/services/project.go
func (p *Project) Update(s *xorm.Session, project *models.Project, u *user.User) (*models.Project, error) {
    // New implementation with proper business logic
}
```

3. **Update model method to delegate**:
```go
// pkg/models/project.go
func (p *Project) Update(s *xorm.Session, a web.Auth) error {
    if ProjectUpdateFunc != nil {
        u, err := user.GetFromAuth(a)
        if err != nil {
            return err
        }
        _, err = ProjectUpdateFunc(s, p, u)
        return err
    }
    return errors.New("ProjectUpdateFunc not initialized")
}
```

### Service Method Structure

Follow this pattern for service methods:

```go
func (s *ServiceType) MethodName(session *xorm.Session, data *models.DataType, user *user.User) (*models.DataType, error) {
    // 1. Permission checks
    can, err := data.CanPerformAction(session, user)
    if err != nil {
        return nil, err
    }
    if !can {
        return nil, &models.ErrGenericForbidden{}
    }

    // 2. Validation
    err = s.validate(session, data)
    if err != nil {
        return nil, err
    }

    // 3. Business logic
    // ... complex operations, cascading updates, etc.

    // 4. Database operations
    _, err = session.ID(data.ID).Update(data)
    if err != nil {
        return nil, err
    }

    // 5. Event dispatching (if needed)
    err = events.Dispatch(&models.SomeEvent{Data: data, User: user})
    if err != nil {
        return nil, err
    }

    return data, nil
}
```

---

## Migration Checklist

When refactoring a model method to services:

- [ ] Create service method with full business logic
- [ ] Add dependency inversion function variable in models
- [ ] Set function variable in services `init()`
- [ ] Update model method to delegate to service
- [ ] Move complex tests from models to services
- [ ] Keep simple validation tests in models
- [ ] Add deprecation comments to old model functions
- [ ] Verify no circular import issues
- [ ] Test both packages independently
- [ ] Update any handlers to use services directly

---

This guide should be updated as new patterns emerge during the refactoring process. Always prioritize clean architecture and testability over convenience.
