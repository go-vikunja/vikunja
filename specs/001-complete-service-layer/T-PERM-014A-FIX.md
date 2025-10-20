# T-PERM-014A-FIX: Wire TaskComment Permission Function Pointers

**Date**: 2025-01-13  
**Status**: ✅ COMPLETE  
**Priority**: MEDIUM  
**Estimated Time**: 0.25 days  
**Actual Time**: 0.25 days  
**Completion Date**: 2025-01-14  
**Parent Task**: T-PERM-014A Phase 1  

## Problem

During T-PERM-014A Phase 1, permission delegation methods were added to `TaskComment` model, but the function pointers were never wired in `InitCommentService()`. This causes baseline tests to fail.

**Failing Tests**:
- `TestPermissionBaseline_TaskComment/CanRead` - All cases fail
- `TestPermissionBaseline_TaskComment/CanUpdate` - All cases fail
- `TestPermissionBaseline_TaskComment/CanDelete` - All cases fail
- `TestPermissionBaseline_TaskComment/CanCreate` - All cases fail

**Error**: `ErrPermissionDelegationNotInitialized` returned because function pointers are nil

## Root Cause

In Phase 1 (T-PERM-013), permission files were deleted before delegation was fully set up. The delegation methods were added to `pkg/models/task_comments.go`:

```go
// pkg/models/task_comments.go lines 430-460
func (tc *TaskComment) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	if CheckTaskCommentReadFunc == nil {
		return false, 0, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTaskCommentReadFunc(s, tc.ID, a)
}

func (tc *TaskComment) CanWrite(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckTaskCommentWriteFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTaskCommentWriteFunc(s, tc.ID, a)
}

func (tc *TaskComment) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckTaskCommentUpdateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTaskCommentUpdateFunc(s, tc.ID, a)
}

func (tc *TaskComment) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckTaskCommentDeleteFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTaskCommentDeleteFunc(s, tc.ID, a)
}

func (tc *TaskComment) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckTaskCommentCreateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTaskCommentCreateFunc(s, tc.ID, a)
}
```

**But the function pointers are never initialized** in `pkg/services/comment.go`'s `InitCommentService()`.

## Solution

Wire the function pointers in `InitCommentService()` to delegate to `CommentService` methods:

```go
// pkg/services/comment.go - Add to InitCommentService()

models.CheckTaskCommentReadFunc = func(s *xorm.Session, commentID int64, a web.Auth) (bool, int, error) {
	cs := NewCommentService(s.Engine())
	// Need to get comment to find taskID
	comment, err := cs.GetByID(s, commentID)
	if err != nil {
		return false, 0, err
	}
	return cs.CanRead(s, comment.TaskID, a)
}

models.CheckTaskCommentWriteFunc = func(s *xorm.Session, commentID int64, a web.Auth) (bool, error) {
	cs := NewCommentService(s.Engine())
	comment, err := cs.GetByID(s, commentID)
	if err != nil {
		return false, err
	}
	// Write is same as Update for comments
	return cs.CanUpdate(s, commentID, comment.TaskID, a)
}

models.CheckTaskCommentUpdateFunc = func(s *xorm.Session, commentID int64, a web.Auth) (bool, error) {
	cs := NewCommentService(s.Engine())
	comment, err := cs.GetByID(s, commentID)
	if err != nil {
		return false, err
	}
	return cs.CanUpdate(s, commentID, comment.TaskID, a)
}

models.CheckTaskCommentDeleteFunc = func(s *xorm.Session, commentID int64, a web.Auth) (bool, error) {
	cs := NewCommentService(s.Engine())
	comment, err := cs.GetByID(s, commentID)
	if err != nil {
		return false, err
	}
	return cs.CanDelete(s, commentID, comment.TaskID, a)
}

models.CheckTaskCommentCreateFunc = func(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	cs := NewCommentService(s.Engine())
	return cs.CanCreate(s, taskID, a)
}
```

**Note**: CanCreate uses taskID directly (doesn't need commentID), others need to fetch comment first to get taskID.

## Files to Modify

1. **pkg/services/comment.go** - Add function pointer wiring to `InitCommentService()`
2. **pkg/models/task_comments.go** - Verify function pointer declarations exist

## Verification

```bash
cd /home/aron/projects/vikunja
export VIKUNJA_SERVICE_ROOTPATH=$(pwd)

# Test TaskComment permissions
go test ./pkg/services -run "TestPermissionBaseline_TaskComment" -v

# Should show all passing:
# ✅ TestPermissionBaseline_TaskComment/CanRead
# ✅ TestPermissionBaseline_TaskComment/CanUpdate
# ✅ TestPermissionBaseline_TaskComment/CanDelete
# ✅ TestPermissionBaseline_TaskComment/CanCreate
```

## Success Criteria

- ✅ Function pointers wired in InitCommentService()
- ✅ All TaskComment baseline tests pass
- ✅ No errors about "ErrPermissionDelegationNotInitialized"
- ✅ Production code compiles cleanly

## Dependencies

- Depends on: T-PERM-014A Phase 1 (already complete)
- Blocks: Full baseline test success (currently 4/6 suites passing, should be 5/6 after this)

## Priority Justification

MEDIUM - This is a cleanup task that fixes pre-existing issue from Phase 1. It doesn't block Phase 2 or Phase 3 work, but should be completed before declaring T-PERM-014A fully complete.
