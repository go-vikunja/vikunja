# T-PERM-014A-CIRCULAR-FIX: Fix Circular Dependency Regressions

**Parent**: [T-PERMISSIONS-TASKS-PART3.md](./T-PERMISSIONS-TASKS-PART3.md) - T-PERM-014A  
**Status**: ✅ **COMPLETE** (2025-01-14)  
**Completion Time**: 0.5 days  
**Type**: Critical Regression Fix

---

## Problem Statement

Phase 3 service layer refactoring introduced **critical circular dependencies** in service constructors that caused infinite recursion and stack overflow during test execution:

### Circular Dependency Cycles

1. **Cycle 1**: `NewProjectService` → `NewLinkShareService` → `NewProjectService` → ...
   - Location: `pkg/services/project.go:96` → `pkg/services/link_share.go:123`
   - Impact: Any code creating ProjectService would crash immediately

2. **Cycle 2**: `NewTaskService` → `NewReactionsService` → `NewCommentService` → `NewTaskService` → ...
   - Location: `pkg/services/task.go:701` → `pkg/services/reactions.go:37` → `pkg/services/comment.go:42`
   - Impact: Any code creating TaskService would crash immediately

### Symptoms

```
fatal error: stack overflow

runtime stack:
runtime.throw({0x1435cd4?, 0x0?})
        /usr/local/go/src/runtime/panic.go:1094 +0x48

goroutine 1 [running]:
code.vikunja.io/api/pkg/services.NewProjectService(0x0)
        /home/aron/projects/vikunja/pkg/services/project.go:96
code.vikunja.io/api/pkg/services.NewLinkShareService(...)
        /home/aron/projects/vikunja/pkg/services/link_share.go:123
code.vikunja.io/api/pkg/services.NewProjectService(0x0)
        /home/aron/projects/vikunja/pkg/services/project.go:96
[repeats until stack overflow]
```

**Result**: All baseline permission tests failed to run (timeout/crash before test execution)

---

## Root Cause

Phase 3 changes added cross-service dependencies without considering the dependency graph:

**Before Phase 3** (Working):
- Services were mostly independent or had one-way dependencies
- No circular references in constructors

**After Phase 3** (Broken):
- `LinkShareService` added `ProjectService` dependency (for `GetByIDSimple()` calls)
- `ProjectService` already had `LinkShareService` dependency
- `CommentService` added `TaskService` dependency (for `GetByIDSimple()` call)
- `TaskService` already had `CommentService` dependency (via `ReactionsService`)

---

## Solution: Lazy Initialization Pattern

Instead of eagerly initializing dependencies in constructors, use **lazy initialization** to break cycles:

### Pattern Implementation

```go
type ServiceWithDependency struct {
    DB              *xorm.Engine
    DependentService *DependentService  // May be nil initially
}

// Lazy initializer - creates dependency only when first needed
func (s *ServiceWithDependency) getDependentService() *DependentService {
    if s.DependentService == nil {
        s.DependentService = NewDependentService(s.DB)
    }
    return s.DependentService
}

// Constructor leaves circular dependency nil
func NewServiceWithDependency(db *xorm.Engine) *ServiceWithDependency {
    return &ServiceWithDependency{
        DB:              db,
        DependentService: nil,  // Lazily initialized to avoid circular dependency
    }
}

// Usage - call lazy initializer instead of accessing field directly
func (s *ServiceWithDependency) SomeMethod(session *xorm.Session, id int64) error {
    dependency, err := s.getDependentService().SomeOtherMethod(session, id)
    // ...
}
```

---

## Changes Made

### 1. LinkShareService (Cycle 1 Fix)

**File**: `pkg/services/link_share.go`

**Changes**:
1. Updated `NewLinkShareService()` to set `ProjectService: nil`
2. Added `getProjectService()` lazy initializer method
3. Updated 3 usages to call `lss.getProjectService().GetByIDSimple()`:
   - `GetByProjectIDWithOptions()` (line 245)
   - `canDoLinkShare()` (line 423)
   - `GetProjectByShareHash()` (line 472)

**Before**:
```go
func NewLinkShareService(engine *xorm.Engine) *LinkShareService {
    return &LinkShareService{
        DB:             engine,
        ProjectService: NewProjectService(engine),  // ❌ Circular dependency
    }
}
```

**After**:
```go
// getProjectService lazily initializes ProjectService if nil (avoids circular dependency)
func (lss *LinkShareService) getProjectService() *ProjectService {
    if lss.ProjectService == nil {
        lss.ProjectService = NewProjectService(lss.DB)
    }
    return lss.ProjectService
}

func NewLinkShareService(engine *xorm.Engine) *LinkShareService {
    return &LinkShareService{
        DB:             engine,
        ProjectService: nil,  // ✅ Lazily initialized to avoid circular dependency
    }
}
```

### 2. CommentService (Cycle 2 Fix - Part 1)

**File**: `pkg/services/comment.go`

**Changes**:
1. Updated `NewCommentService()` to set `TaskService: nil`
2. Added `getTaskService()` lazy initializer method
3. Updated 1 usage to call `cs.getTaskService().GetByIDSimple()`:
   - `Delete()` method (line 493)

**Before**:
```go
func NewCommentService(db *xorm.Engine) *CommentService {
    return &CommentService{
        DB:               db,
        LinkShareService: NewLinkShareService(db),
        TaskService:      NewTaskService(db),  // ❌ Circular dependency
    }
}
```

**After**:
```go
// getTaskService lazily initializes TaskService if nil (avoids circular dependency)
func (cs *CommentService) getTaskService() *TaskService {
    if cs.TaskService == nil {
        cs.TaskService = NewTaskService(cs.DB)
    }
    return cs.TaskService
}

func NewCommentService(db *xorm.Engine) *CommentService {
    return &CommentService{
        DB:               db,
        LinkShareService: NewLinkShareService(db),
        TaskService:      nil,  // ✅ Lazily initialized to avoid circular dependency
    }
}
```

### 3. ReactionsService (Cycle 2 Fix - Part 2)

**File**: `pkg/services/reactions.go`

**Changes**:
1. Updated `NewReactionsService()` to set `CommentService: nil`
2. Added `getCommentService()` lazy initializer method
3. No direct usages (CommentService only used internally if needed)

**Before**:
```go
func NewReactionsService(db *xorm.Engine) *ReactionsService {
    return &ReactionsService{
        DB:             db,
        CommentService: NewCommentService(db),  // ❌ Circular dependency
    }
}
```

**After**:
```go
// getCommentService lazily initializes CommentService if nil (avoids circular dependency)
func (rs *ReactionsService) getCommentService() *CommentService {
    if rs.CommentService == nil {
        rs.CommentService = NewCommentService(rs.DB)
    }
    return rs.CommentService
}

func NewReactionsService(db *xorm.Engine) *ReactionsService {
    return &ReactionsService{
        DB:             db,
        CommentService: nil,  // ✅ Lazily initialized to avoid circular dependency
    }
}
```

---

## Verification

### Test Results

**Before Fix**: Stack overflow, 0/6 tests passing
```bash
$ VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test -v ./pkg/services -run 'TestPermissionBaseline_Project'
fatal error: stack overflow
```

**After Fix**: All tests passing ✅
```bash
$ VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test -v ./pkg/services -run 'TestPermissionBaseline.*'
=== RUN   TestPermissionBaseline_Project
--- PASS: TestPermissionBaseline_Project (0.13s)
=== RUN   TestPermissionBaseline_Task
--- PASS: TestPermissionBaseline_Task (0.09s)
=== RUN   TestPermissionBaseline_LinkSharing
--- PASS: TestPermissionBaseline_LinkSharing (0.04s)
=== RUN   TestPermissionBaseline_Label
--- PASS: TestPermissionBaseline_Label (0.04s)
=== RUN   TestPermissionBaseline_TaskComment
--- PASS: TestPermissionBaseline_TaskComment (0.04s)
=== RUN   TestPermissionBaseline_Subscription
--- PASS: TestPermissionBaseline_Subscription (0.02s)
PASS
ok      code.vikunja.io/api/pkg/services        0.391s
```

**Final Status**: ✅ 6/6 baseline tests passing (100%)

### Compilation

```bash
$ go build ./pkg/...
[Success - no errors]
```

---

## Impact Analysis

### Files Modified

1. **pkg/services/link_share.go**
   - Added lazy initializer: `getProjectService()`
   - Updated 3 call sites
   - Constructor updated

2. **pkg/services/comment.go**
   - Added lazy initializer: `getTaskService()`
   - Updated 1 call site
   - Constructor updated

3. **pkg/services/reactions.go**
   - Added lazy initializer: `getCommentService()`
   - No call site changes needed
   - Constructor updated

### Architecture Benefits

1. **Breaking Cycles**: Lazy initialization inherently breaks circular dependencies by deferring creation
2. **Thread Safety**: Not required here (services created once at startup)
3. **Performance**: Negligible overhead (null check + one-time initialization)
4. **Maintainability**: Clear pattern for handling future circular dependencies

---

## Lessons Learned

### Why This Happened

Phase 3 added service dependencies **without analyzing the dependency graph**:
- Added `LinkShareService.ProjectService` without checking if `ProjectService` depends on `LinkShareService` ✅ (it did)
- Added `CommentService.TaskService` without checking the full dependency chain ✅ (TaskService → ReactionsService → CommentService)

### Prevention Strategy

**Before adding a new service dependency**:
1. Check if the target service already depends (directly or indirectly) on your service
2. Visualize the dependency graph if unsure
3. Consider lazy initialization for cross-cutting dependencies (like ProjectService, TaskService)

### When to Use Lazy Initialization

✅ **Use lazy initialization when**:
- Adding a dependency that might create a cycle
- Dependency is only needed in a few methods (not critical path)
- Service is "central" (like ProjectService, TaskService) and likely to have many dependents

❌ **Don't use lazy initialization when**:
- Dependency is always needed (eager initialization is clearer)
- Service has no risk of circular dependencies
- Performance-critical code path (though overhead is minimal)

---

## Related Documents

- 📋 [T-PERMISSIONS-TASKS-PART3.md](./T-PERMISSIONS-TASKS-PART3.md) - Parent task (T-PERM-014A)
- 📄 [T-PERM-014A-PHASE3-PLAN.md](./T-PERM-014A-PHASE3-PLAN.md) - Original Phase 3 plan
- 📋 [T-PERMISSIONS-TEST-CHECKLIST.md](./T-PERMISSIONS-TEST-CHECKLIST.md) - Test verification status

---

## Completion Checklist

- ✅ Identified all circular dependencies (2 cycles found)
- ✅ Implemented lazy initialization for `LinkShareService.ProjectService`
- ✅ Implemented lazy initialization for `CommentService.TaskService`
- ✅ Implemented lazy initialization for `ReactionsService.CommentService`
- ✅ Updated all call sites (4 total)
- ✅ Verified compilation succeeds
- ✅ Verified all 6 baseline tests pass (100%)
- ✅ Documented pattern for future reference
- ✅ Updated task tracking documents

**Status**: ✅ **COMPLETE** (2025-01-14)
