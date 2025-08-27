# Vikunja Service Layer Refactoring Guide

This document serves as the **source of truth** for all refactoring work in the Vikunja project. It documents the core architectural patterns, dependency management strategies, and testing philosophies established during the service layer refactor.

> ## The Golden Rule: Move Logic, Don't Expose It
>
> This is the most important rule of the entire refactor.
>
> When a service needs logic that is currently inside a private `models` function, the solution is **always** to **MOVE** the logic from the model into the service.
>
> You must **NEVER** solve this problem by making the model function public and calling it from the service. This violates the architecture and re-introduces the problems we are trying to fix.
>
> * **Correct ✅:** Copying the permission-checking code from a model method into a new service method.
> * **Incorrect ❌:** Making a model's `CanRead()` or `CanWrite()` method public and calling it from a service.

## Table of Contents

1.  [The Core Architecture ("Chef, Waiter, Pantry")](#1-the-core-architecture-chef-waiter-pantry)
2.  [Key Patterns & Implementation](#2-key-patterns--implementation)
      * [A. Handler Wrappers (Eliminating Boilerplate)](#a-handler-wrappers-eliminating-boilerplate)
      * [B. Dependency Inversion (For Backward Compatibility)](#b-dependency-inversion-for-backward-compatibility)
3.  [The Testing Philosophy](#3-the-testing-philosophy)

---

## 1\. The Core Architecture ("Chef, Waiter, Pantry")

Our architecture separates concerns into three distinct layers:

### 🧑‍🍳 Services (`pkg/services`) - The "Chef"
    This is the "brains" of the application. It contains **ALL** business logic, including permission checks, validation, and orchestrating complex operations. Services are decoupled from the web layer and use `*user.User`.

### 🧑‍💼 Handlers (`pkg/routes`) - The "Waiter"
    This is a thin "glue" layer. Its only job is to parse HTTP requests, call the appropriate service method, and return an HTTP response. We use **handler wrappers** to eliminate boilerplate.

### 🏪 Models (`pkg/models`) - The "Pantry"
    This is a "dumb" data layer. Its only job is to define data structures and provide basic, low-level database access (CRUD) functions. Models **must not** contain business logic or permission checks.

---

## 2\. Key Patterns & Implementation

### A. Handler Wrappers (Eliminating Boilerplate)

To keep handlers clean, we use a wrapper function. This is the **required pattern** for all new handlers.

**The Wrapper (`pkg/web/handler/wrapper.go`):**

```go
// WithDBAndUser handles session creation, user auth, and error handling.
func WithDBAndUser(logicFunc func(s *xorm.Session, u *user.User, c echo.Context) error, needsTransaction bool) echo.HandlerFunc {
    return func(c echo.Context) error {
        // ... boilerplate for session, user, transaction, and error handling ...
        return logicFunc(s, u, c)
    }
}
```

**Usage (in `pkg/routes/api/v1/...`):**

```go
// The handler is now a simple, clean function containing only business logic.
func getProject(s *xorm.Session, u *user.User, c echo.Context) error {
    // ... parse params, call service, return response ...
}

// The route registration uses the wrapper.
a.GET("/projects/:id", handler.WithDBAndUser(getProject, false))
```

### B. Dependency Inversion (For Backward Compatibility)

When a refactored service needs to be called by an old, deprecated model method, we must break the `models` -\> `services` import cycle.

**The Solution:** Use a function variable in the model, set by the service's `init()` function.

**1. In the Model (`pkg/models/tasks.go`):**

```go
// 1. Define a public function variable "placeholder".
var TaskUpdateFunc func(s *xorm.Session, t *Task, u *user.User) error

// 2. The deprecated method just calls the placeholder.
// @Deprecated: Use services.TaskService.Update instead.
func (t *Task) Update(s *xorm.Session, u *user.User) error {
    if TaskUpdateFunc != nil {
        return TaskUpdateFunc(s, t, u)
    }
    return errors.New("TaskUpdateFunc not initialized")
}
```

**2. In the Service (`pkg/services/task.go`):**

```go
// 3. The service's init() function "plugs in" the real implementation.
func init() {
    models.TaskUpdateFunc = NewTaskService(nil).Update
}
```

-----

## 3\. The Testing Philosophy

Our testing strategy is layered to match the architecture.

  * **Model Tests (`pkg/models/`):**
    Simple **unit tests** only. They should verify struct tags, defaults, and very basic database interactions. They must not have complex dependencies.

  * **Service Tests (`pkg/services/`):**
    This is the home for our **integration tests**. These tests verify the complete business logic, including permission checks, side effects (like event dispatching), and interactions between different models.

  * **Test Environment Setup:**
    Tests require a specific setup to run correctly.

    1.  **Test Prerequisites (Building the Frontend):** Some test suites, especially the end-to-end web tests (`mage test:web`), require the frontend assets to be built first. If these are missing, you may see errors like `pattern dist: no matching files found`. To fix this, run the following commands from the project root **once** at the beginning of your task:

        ```bash
        cd frontend
        pnpm install
        pnpm run build
        cd ..
        ```

    2.  **Run Command:** Always run tests using the `mage` commands (`mage test:feature` for backend, `mage test:web` for end-to-end) or by setting the environment variable: `VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test ./...`

    3.  **`TestMain` for Setup:** Each test package (`models`, `services`, etc.) should have a `main_test.go` file with a `TestMain` function that initializes the database and other dependencies (`models.SetupTests()`, `user.InitTests()`, etc.).

    4.  **The "Master Switch" (`pkg/testutil`):** For test suites that need the dependency inversion to work (like the `models` tests), the `main_test.go` file must contain a blank import to our test utility package: `_ "code.vikunja.io/api/pkg/testutil"`. This safely triggers the `init()` functions from the `services` package.
