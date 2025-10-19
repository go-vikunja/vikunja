# T-PERMISSIONS Test Completion Checklist

**Purpose**: Ensure ALL tests pass before T-PERMISSIONS refactor is considered complete  
**Last Updated**: 2025-01-14  
**Status**: ✅ COMPLETE - 6/6 baseline tests passing  

## Test Completion Requirements

Per **T-PERM-017** (Final Verification & Documentation), all baseline permission tests MUST pass before T-PERMISSIONS is complete.

## Current Test Status

### Baseline Permission Tests: 6/6 Passing ✅

| # | Test Suite | Status | Issue | Fix Task | Completion |
|---|------------|--------|-------|----------|------------|
| 1 | TestPermissionBaseline_Project | ✅ PASS | None | N/A | - |
| 2 | TestPermissionBaseline_Task | ✅ PASS | None | N/A | - |
| 3 | TestPermissionBaseline_LinkSharing | ✅ PASS | None | N/A | - |
| 4 | TestPermissionBaseline_Label | ✅ PASS | None | N/A | - |
| 5 | TestPermissionBaseline_TaskComment | ✅ PASS | Fixed | **T-PERM-014A-FIX** | 2025-01-14 |
| 6 | TestPermissionBaseline_Subscription | ✅ PASS | Fixed | **T-PERM-014A-SUBSCRIPTION-FIX** | 2025-01-14 |

**Passing Rate**: 6/6 (100%) ✅  
**Target**: 6/6 (100%) ✅  
**Status**: ALL TESTS PASSING

## ✅ Completed Fix Tasks

### 1. T-PERM-014A-FIX: Wire TaskComment Permission Function Pointers
**Status**: ✅ COMPLETE  
**Priority**: HIGH  
**Time**: 0.25 days  
**Completion Date**: 2025-01-14  
**File**: [T-PERM-014A-FIX.md](./T-PERM-014A-FIX.md)  

**Issue**: TaskComment delegation methods exist but function pointers not wired in InitCommentService()

**Fixed Tests**:
- TestPermissionBaseline_TaskComment/CanRead (all cases) ✅
- TestPermissionBaseline_TaskComment/CanUpdate (all cases) ✅
- TestPermissionBaseline_TaskComment/CanDelete (all cases) ✅
- TestPermissionBaseline_TaskComment/CanCreate (all cases) ✅

**Fix Applied**: Added function pointer wiring to `pkg/services/comment.go`:
```go
models.CheckTaskCommentReadFunc = func(s *xorm.Session, commentID int64, a web.Auth) (bool, int, error) {
    cs := NewCommentService(s.Engine())
    comment := &models.TaskComment{ID: commentID}
    err := cs.getTaskCommentSimple(s, comment)
    if err != nil {
        return false, 0, err
    }
    return cs.CanRead(s, comment.TaskID, a)
}
// ... (4 more function pointers)
```

**Result**: TaskComment baseline tests now passing ✅

---

### 2. T-PERM-014A-SUBSCRIPTION-FIX: Fix Subscription Baseline Test
**Status**: ✅ COMPLETE  
**Priority**: HIGH  
**Time**: 0.25 days  
**Completion Date**: 2025-01-14  
**File**: [T-PERM-014A-SUBSCRIPTION-FIX.md](./T-PERM-014A-SUBSCRIPTION-FIX.md)  

**Issue**: Test creates Subscription with `Entity` string field instead of `EntityType` enum field

**Fixed Tests**:
- TestPermissionBaseline_Subscription/CanCreate/UserWithTaskReadPermission_CanSubscribe ✅
- TestPermissionBaseline_Subscription/CanCreate/UserWithoutTaskPermission_CannotSubscribe ✅
- TestPermissionBaseline_Subscription/CanDelete/SubscriptionOwner_CanUnsubscribe ✅

**Fix Applied**: Updated `pkg/services/permissions_baseline_test.go` (4 locations):
```go
// BEFORE
subscription := &models.Subscription{
    Entity:     "task",  // ❌ Wrong field
    EntityID:   taskID,
}

// AFTER
subscription := &models.Subscription{
    EntityType: models.SubscriptionEntityTask,  // ✅ Correct enum
    EntityID:   taskID,
}
```

**Result**: Subscription baseline tests now passing ✅

---

### 3. T-PERM-014A-CIRCULAR-FIX: Fix Circular Dependency Regressions
**Status**: ✅ COMPLETE  
**Priority**: CRITICAL  
**Time**: 0.5 days  
**Completion Date**: 2025-01-14  
**File**: [T-PERM-014A-CIRCULAR-FIX.md](./T-PERM-014A-CIRCULAR-FIX.md)  

**Issue**: Phase 3 service refactoring introduced circular dependencies causing stack overflow

**Circular Dependency Cycles**:
1. `NewProjectService` → `NewLinkShareService` → `NewProjectService` → ...
2. `NewTaskService` → `NewReactionsService` → `NewCommentService` → `NewTaskService` → ...

**Impact**: All tests failed with stack overflow before execution could begin (0/6 passing)

**Fix Applied**: Implemented lazy initialization pattern in 3 service files:

1. **pkg/services/link_share.go**
   - Added `getProjectService()` lazy initializer
   - Updated 3 call sites to use lazy initializer
   - Constructor sets `ProjectService: nil`

2. **pkg/services/comment.go**
   - Added `getTaskService()` lazy initializer
   - Updated 1 call site to use lazy initializer
   - Constructor sets `TaskService: nil`

3. **pkg/services/reactions.go**
   - Added `getCommentService()` lazy initializer
   - No call site updates needed
   - Constructor sets `CommentService: nil`

**Result**: All 6 baseline tests now pass (0/6 → 6/6) ✅

---

## Verification Commands
subscription := &models.Subscription{
    Entity:   "task",  // ❌ Wrong field
    EntityID: 1,
}

// AFTER
subscription := &models.Subscription{
    EntityType: models.SubscriptionEntityTask,  // ✅ Correct field
    EntityID:   1,
}
```

**Result**: Subscription baseline tests now passing ✅

---

## Implementation Summary

Both fixes were implemented in parallel on 2025-01-14:
1. T-PERM-014A-FIX: Wired TaskComment function pointers (0.25 days)
2. T-PERM-014A-SUBSCRIPTION-FIX: Fixed Subscription test field (0.25 days)

**Total Time**: 0.25 days (parallel execution)  
**Final Result**: 6/6 baseline tests passing ✅

---

## Required Fix Tasks (ARCHIVED - NOW COMPLETE)

### 1. T-PERM-014A-FIX: Wire TaskComment Permission Function Pointers
**Status**: ✅ COMPLETE (2025-01-14)  
**Priority**: HIGH  
**Time**: 0.25 days  
**File**: [T-PERM-014A-FIX.md](./T-PERM-014A-FIX.md)  

**Issue**: TaskComment delegation methods exist but function pointers not wired in InitCommentService()

**Failing Tests**:
- TestPermissionBaseline_TaskComment/CanRead (all cases)
- TestPermissionBaseline_TaskComment/CanUpdate (all cases)
- TestPermissionBaseline_TaskComment/CanDelete (all cases)
- TestPermissionBaseline_TaskComment/CanCreate (all cases)

**Fix**: Add function pointer wiring to `pkg/services/comment.go`:
```go
models.CheckTaskCommentReadFunc = func(s *xorm.Session, commentID int64, a web.Auth) (bool, int, error) {
    cs := NewCommentService(s.Engine())
    comment, err := cs.GetByID(s, commentID)
    if err != nil {
        return false, 0, err
    }
    return cs.CanRead(s, comment.TaskID, a)
}
// ... (4 more function pointers)
```

**Expected Result**: 5/6 tests passing

---

### 2. T-PERM-014A-SUBSCRIPTION-FIX: Fix Subscription Baseline Test
**Status**: TODO  
**Priority**: HIGH  
**Time**: 0.25 days  
**File**: [T-PERM-014A-SUBSCRIPTION-FIX.md](./T-PERM-014A-SUBSCRIPTION-FIX.md)  

**Issue**: Test creates Subscription with `Entity` string field instead of `EntityType` enum field

**Failing Tests**:
- TestPermissionBaseline_Subscription/CanCreate/UserWithTaskReadPermission_CanSubscribe
- TestPermissionBaseline_Subscription/CanCreate/UserWithoutTaskPermission_CannotSubscribe
- TestPermissionBaseline_Subscription/CanDelete/SubscriptionOwner_CanUnsubscribe

**Fix**: Update `pkg/services/permissions_baseline_test.go` (3-4 locations):
```go
// BEFORE
subscription := &models.Subscription{
    Entity:   "task",  // ❌ Wrong field
    EntityID: 1,
    UserID:   1,
}

// AFTER
subscription := &models.Subscription{
    EntityType: models.SubscriptionEntityTask,  // ✅ Correct field
    EntityID:   1,
    UserID:     1,
}
```

**Expected Result**: 6/6 tests passing (after T-PERM-014A-FIX also completed)

---

## Implementation Order

### Option A: Sequential (Safer)
1. Complete T-PERM-014A-FIX first (0.25 days)
2. Verify 5/6 tests passing
3. Complete T-PERM-014A-SUBSCRIPTION-FIX (0.25 days)
4. Verify 6/6 tests passing
5. **Total time**: 0.5 days

### Option B: Parallel (Faster) ⭐ RECOMMENDED
1. Fix both issues simultaneously (0.25 days)
2. Verify 6/6 tests passing immediately
3. **Total time**: 0.25 days

**Recommendation**: Use Option B - both fixes are independent and simple

---

## Verification Commands

### Quick Check (Current Status)
```bash
cd /home/aron/projects/vikunja
export VIKUNJA_SERVICE_ROOTPATH=$(pwd)
go test ./pkg/services -run "TestPermissionBaseline" -v | grep -E "^(PASS|FAIL|---)"
```

### After T-PERM-014A-FIX
```bash
go test ./pkg/services -run "TestPermissionBaseline_TaskComment" -v
# Expected: All TaskComment tests PASS
# Status: 5/6 baseline suites passing
```

### After T-PERM-014A-SUBSCRIPTION-FIX
```bash
go test ./pkg/services -run "TestPermissionBaseline_Subscription" -v
# Expected: All Subscription tests PASS
# Status: 6/6 baseline suites passing ✅
```

### Final Verification (Both Complete)
```bash
# All baseline tests
go test ./pkg/services -run "TestPermissionBaseline" -v
# Expected: ALL PASS (6/6 suites)

# Full service test suite
go test ./pkg/services -v
# Expected: ALL PASS

# Complete project test suite
mage test:all
# Expected: ALL PASS
```

---

## Blocking Relationships

```
T-PERM-017 (Final Verification)
    ⬆️ BLOCKS (requires 6/6 baseline tests passing)
    │
    ├─ T-PERM-014A-FIX (TaskComment)
    │   └─ Fixes 1/2 failing baseline suites
    │
    └─ T-PERM-014A-SUBSCRIPTION-FIX (Subscription)
        └─ Fixes 2/2 failing baseline suites
```

**CRITICAL**: T-PERM-017 cannot proceed until both fix tasks are complete.

---

## Success Criteria

- ✅ **All 6 baseline test suites passing**
- ✅ TestPermissionBaseline_Project: PASS
- ✅ TestPermissionBaseline_Task: PASS
- ✅ TestPermissionBaseline_LinkSharing: PASS
- ✅ TestPermissionBaseline_Label: PASS
- ✅ TestPermissionBaseline_TaskComment: PASS
- ✅ TestPermissionBaseline_Subscription: PASS

**Gate for T-PERMISSIONS Completion**: These fixes are MANDATORY before declaring T-PERMISSIONS complete.

---

## Additional Test Suites

Beyond baseline tests, the following must also pass:

### Service Layer Tests
```bash
go test ./pkg/services -v
```
**Current Status**: ⚠️ Some pre-existing failures (not from refactor)
- TestTaskService_GetAllWithMultipleSortParameters: FAIL (pre-existing, DB error)

**Action**: Document as pre-existing, not blocking (doesn't involve permission logic)

### Model Layer Tests
```bash
go test ./pkg/models -v
```
**Current Status**: Unknown (to be verified in T-PERM-016)

**Action**: Will be addressed in T-PERM-016 (Update Model Tests to Pure Structure Tests)

### Full Test Suite
```bash
mage test:all
```
**Current Status**: To be verified after all refactoring complete

**Action**: Final gate in T-PERM-017

---

## Timeline to 100% Test Passing

| Milestone | Tasks | Time | Cumulative |
|-----------|-------|------|------------|
| **Current State** | 4/6 baseline passing | - | - |
| **Fix TaskComment** | T-PERM-014A-FIX | 0.25 days | 0.25 days |
| **Fix Subscription** | T-PERM-014A-SUBSCRIPTION-FIX | 0.0 days (parallel) | 0.25 days |
| **6/6 Baseline Passing** | ✅ Complete | - | **0.25 days** |
| **Complete Phase 3** | T-PERM-014A-PHASE3 | 1-1.5 days | 1.5-1.75 days |
| **Final Verification** | T-PERM-017 | 0.5 days | 2-2.25 days |

**Total Time to 100% Passing Tests**: 2-2.25 days from current state

---

## Priority Actions

### IMMEDIATE (Today)
1. ⚠️ Complete T-PERM-014A-FIX (0.25 days)
2. ⚠️ Complete T-PERM-014A-SUBSCRIPTION-FIX (parallel)
3. ✅ Verify 6/6 baseline tests passing

### NEXT (This Week)
4. Continue with T-PERM-014A-PHASE3 (high-impact helper removals)
5. Run full verification in T-PERM-017

### BEFORE COMPLETION
- ✅ 6/6 baseline tests passing (MANDATORY)
- ✅ All service tests passing (or pre-existing failures documented)
- ✅ Full test suite clean (mage test:all)

---

## References

- [T-PERM-014A-FIX.md](./T-PERM-014A-FIX.md) - TaskComment function pointer wiring
- [T-PERM-014A-SUBSCRIPTION-FIX.md](./T-PERM-014A-SUBSCRIPTION-FIX.md) - Subscription test field fix
- [T-PERMISSIONS-TASKS-PART3.md](./T-PERMISSIONS-TASKS-PART3.md#t-perm-017-final-verification--documentation) - T-PERM-017 requirements
- [T-PERM-014A-PROGRESS.md](./T-PERM-014A-PROGRESS.md) - Overall refactor progress
