# T-PERM-014A-SUBSCRIPTION-FIX: Fix Subscription Baseline Test

**Date**: 2025-01-13  
**Status**: ✅ COMPLETE  
**Priority**: HIGH (Must complete before T-PERMISSIONS is done)  
**Estimated Time**: 0.25 days  
**Actual Time**: 0.25 days  
**Completion Date**: 2025-01-14  
**Parent Task**: T-PERM-000 (Baseline Permission Tests)  

## Problem

The Subscription baseline tests are failing because the test code uses the wrong field name when creating test subscriptions. The test uses `Entity` (a string field marked as transient) instead of `EntityType` (the enum field that's actually persisted and used by the code).

**Failing Tests**:
- `TestPermissionBaseline_Subscription/CanCreate/UserWithTaskReadPermission_CanSubscribe`
- `TestPermissionBaseline_Subscription/CanCreate/UserWithoutTaskPermission_CannotSubscribe`
- `TestPermissionBaseline_Subscription/CanDelete/SubscriptionOwner_CanUnsubscribe`

**Error Message**:
```
Error: Subscription entity type is unknown [EntityType: 0]
```

## Root Cause

In `pkg/services/permissions_baseline_test.go`, the test creates subscriptions like this:

```go
// WRONG - uses Entity string field
subscription := &models.Subscription{
    Entity:   "task",  // ❌ This is the transient field, not persisted
    EntityID: 1,
    UserID:   1,
}
```

But the `Subscription` model has two fields:
```go
// pkg/models/subscription.go
type Subscription struct {
    // Entity is the entity this subscription refers to (task, project, etc.)
    // This is a computed field from EntityType for API responses
    Entity string `xorm:"-" json:"-"` // Transient, not in DB

    // EntityType represents the type of entity this subscription is for
    EntityType SubscriptionEntityType `xorm:"bigint INDEX not null" json:"entity" param:"entity"`
}
```

The code actually checks `EntityType` (the enum), not `Entity` (the string). The test is setting the wrong field.

## Solution

Update the test to use `EntityType` instead of `Entity`:

### File to Modify: `pkg/services/permissions_baseline_test.go`

**Line ~1122** (in TestPermissionBaseline_Subscription/CanCreate):
```go
// BEFORE
subscription := &models.Subscription{
    Entity:   "task",
    EntityID: 1,
    UserID:   6,
}

// AFTER
subscription := &models.Subscription{
    EntityType: models.SubscriptionEntityTask, // Use the enum constant
    EntityID:   1,
    UserID:     6,
}
```

**Line ~1135** (in TestPermissionBaseline_Subscription/CanCreate second case):
```go
// BEFORE
subscription := &models.Subscription{
    Entity:   "task",
    EntityID: 1,
    UserID:   13,
}

// AFTER
subscription := &models.Subscription{
    EntityType: models.SubscriptionEntityTask,
    EntityID:   1,
    UserID:     13,
}
```

**Line ~1152** (in TestPermissionBaseline_Subscription/CanDelete):
```go
// BEFORE (if it exists - verify first)
subscription := &models.Subscription{
    Entity:   "task",
    EntityID: 1,
    UserID:   1,
}

// AFTER
subscription := &models.Subscription{
    EntityType: models.SubscriptionEntityTask,
    EntityID:   1,
    UserID:     1,
}
```

## Verification

```bash
cd /home/aron/projects/vikunja
export VIKUNJA_SERVICE_ROOTPATH=$(pwd)

# Test Subscription permissions specifically
go test ./pkg/services -run "TestPermissionBaseline_Subscription" -v

# Should show all passing:
# ✅ TestPermissionBaseline_Subscription/CanCreate/UserWithTaskReadPermission_CanSubscribe
# ✅ TestPermissionBaseline_Subscription/CanCreate/UserWithoutTaskPermission_CannotSubscribe
# ✅ TestPermissionBaseline_Subscription/CanDelete/SubscriptionOwner_CanUnsubscribe
# ✅ TestPermissionBaseline_Subscription/CanDelete/OtherUser_CannotUnsubscribe

# Then verify all baseline tests
go test ./pkg/services -run "TestPermissionBaseline" -v
# Should show 5/6 or 6/6 passing (depending on T-PERM-014A-FIX status)
```

## Success Criteria

- ✅ All Subscription baseline test cases pass
- ✅ No "entity type is unknown" errors
- ✅ Tests use correct `EntityType` enum field
- ✅ Combined with T-PERM-014A-FIX, should have 6/6 baseline suites passing

## Additional Notes

### Why This Wasn't Caught Earlier

1. **Baseline tests were created during refactor**: They don't exist in `vikunja_original_main`
2. **Field confusion**: Having both `Entity` (string) and `EntityType` (enum) is confusing
3. **JSON unmarshaling**: The JSON tag on `EntityType` is `"entity"`, which may have led to confusion
4. **Test-only issue**: Production code correctly uses `EntityType`

### Enum Constants Available

From `pkg/models/subscription.go`:
```go
const (
    SubscriptionEntityUnknown SubscriptionEntityType = iota
    SubscriptionEntityTask
    SubscriptionEntityProject
)
```

Use `models.SubscriptionEntityTask` for task subscriptions in tests.

## Priority Justification

**HIGH** - This must be fixed before T-PERMISSIONS refactor is considered complete:
- **T-PERM-017 requirement**: "All baseline tests pass"
- **Quality gate**: Baseline tests are the safety net for the entire refactor
- **Quick fix**: Only 3-4 lines to change, high impact for minimal effort
- **Blocks completion**: T-PERMISSIONS cannot be marked complete with failing tests

## Dependencies

- **Depends on**: T-PERM-000 (Baseline tests exist)
- **Blocks**: T-PERM-017 (Final Verification & Documentation)
- **Related to**: T-PERM-014A-FIX (TaskComment fix - both needed for 6/6 passing)

## Implementation Time

**Estimated**: 0.25 days (15-30 minutes to fix, rest for testing and verification)
- 5 min: Locate all 3-4 instances in test file
- 5 min: Change `Entity: "task"` to `EntityType: models.SubscriptionEntityTask`
- 5 min: Run tests to verify
- 15 min: Full baseline test suite verification
