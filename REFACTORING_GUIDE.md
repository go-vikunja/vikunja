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

### **A. Handler Wrappers (Eliminating Boilerplate)**

This is the **required pattern** for all new handlers.

// The wrapper in pkg/web/handler/wrapper.go  
func WithDBAndUser(logicFunc func(s \*xorm.Session, u \*user.User, c echo.Context) error, needsTransaction bool) echo.HandlerFunc { /\* ... \*/ }

// Usage in a route file  
func getProject(s \*xorm.Session, u \*user.User, c echo.Context) error { /\* ... \*/ }  
a.GET("/projects/:id", handler.WithDBAndUser(getProject, false))

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

* **Model Tests (pkg/models/):** Simple **unit tests** only.  
* **Service Tests (pkg/services/):** This is the home for our complex **integration tests**.

### **Test Setup Components**

* **TestMain for Setup:** Each test package must have a main\_test.go with a TestMain function that initializes the environment (models.SetupTests(), user.InitTests(), etc.).  
* **The "Master Switch" (pkg/testutil):** For test suites that need dependency inversion to work (like models tests), the TestMain function must explicitly call testutil.Init(). This deterministically sets up service dependencies. **Do not use blank imports.**  
  // In a main\_test.go file  
  import "code.vikunja.io/api/pkg/testutil"

  func TestMain(m \*testing.M) {  
      // ... other setup ...  
      testutil.Init() // Initialize service dependency injection  
      // ... rest of setup ...  
  }  
