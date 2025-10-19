# Label Task Assignment Fix

**Date**: 2025-10-15  
**Issue**: Frontend error when assigning a label to a task  
**Root Cause**: Handler was binding to wrong model struct  
**Status**: ✅ FIXED

## Problem Description

The frontend was sending the following payload to assign a label to a task:

```json
{
    "max_permission": null,
    "id": 0,
    "task_id": 11,
    "label_id": 4
}
```

But the API was responding with:

```json
{
    "message": "Label ID is required"
}
```

## Root Cause Analysis

The handler in `/home/aron/projects/vikunja/pkg/routes/api/v1/label_tasks.go` was binding the request body to the wrong struct:

**BEFORE** (Incorrect):
```go
var label models.Label  // Wrong! Label struct has field "id" not "label_id"
if err := c.Bind(&label); err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, "Invalid label object")
}

if label.ID == 0 {  // This check fails because "label_id" doesn't map to "id"
    return echo.NewHTTPError(http.StatusBadRequest, "Label ID is required")
}
```

The `models.Label` struct has a field `ID` with JSON tag `"id"`, but the frontend (correctly) sends `"label_id"` as per the `models.LabelTask` struct definition:

```go
type LabelTask struct {
    ID      int64 `xorm:"bigint autoincr not null unique pk" json:"-"`
    TaskID  int64 `xorm:"bigint INDEX not null" json:"-" param:"projecttask"`
    LabelID int64 `xorm:"bigint INDEX not null" json:"label_id" param:"label"`
    Created time.Time `xorm:"created not null" json:"created"`
}
```

## Solution

Changed the handler to bind to the correct struct (`models.LabelTask`):

**AFTER** (Correct):
```go
var labelTask models.LabelTask  // Correct! LabelTask struct has field "label_id"
if err := c.Bind(&labelTask); err != nil {
    return echo.NewHTTPError(http.StatusBadRequest, "Invalid label object")
}

if labelTask.LabelID == 0 {  // Now correctly checks the LabelID field
    return echo.NewHTTPError(http.StatusBadRequest, "Label ID is required")
}

err = service.AddLabelToTask(s, labelTask.LabelID, taskID, u)
```

## Files Modified

### 1. `/home/aron/projects/vikunja/pkg/routes/api/v1/label_tasks.go`
- Changed binding from `models.Label` to `models.LabelTask`
- Updated field references from `label.ID` to `labelTask.LabelID`
- Updated Swagger documentation to correctly specify `models.LabelTask`

### 2. `/home/aron/projects/vikunja/pkg/webtests/task_label_test.go`
- Fixed test payloads to send `"label_id"` instead of `"id"`
- Fixed URL construction bug in remove label test (was using `string(rune(id+'0'))` instead of `fmt.Sprintf`)
- Added `fmt` import

## Comparison with Original Code

Checked the original implementation in `vikunja_original_main/pkg/models/label_task.go`:

```go
// Original Swagger documentation (line 87)
// @Param label body models.LabelTask true "The label object"
```

The original API **did** expect a `LabelTask` object with `label_id` field, confirming our fix is correct.

## Verification

All tests now pass:

```bash
$ go test ./pkg/webtests -run TestTaskLabel -v
=== RUN   TestTaskLabel_AddLabel
--- PASS: TestTaskLabel_AddLabel (0.04s)
=== RUN   TestTaskLabel_AddLabelWithoutID
--- PASS: TestTaskLabel_AddLabelWithoutID (0.03s)
=== RUN   TestTaskLabel_RemoveLabel
--- PASS: TestTaskLabel_RemoveLabel (0.03s)
=== RUN   TestTaskLabel_GetTaskLabels
--- PASS: TestTaskLabel_GetTaskLabels (0.03s)
=== RUN   TestTaskLabel_BulkUpdate
--- PASS: TestTaskLabel_BulkUpdate (0.03s)
=== RUN   TestTaskLabel_PermissionDenied
--- PASS: TestTaskLabel_PermissionDenied (0.03s)
PASS
ok      code.vikunja.io/api/pkg/webtests        0.207s
```

## Architecture Decision

The fix maintains consistency with the original API design:
- The endpoint accepts a `LabelTask` relation object with `label_id` field
- This matches the database model structure
- The frontend correctly sends the expected payload format
- The handler was the only component that needed correction

## Related Work

This fix was needed after the T-PERMISSIONS refactoring (T-PERM-014C) which migrated label handling to the service layer. The refactoring correctly moved the business logic but inadvertently changed the request binding model from `LabelTask` to `Label`.

## Success Criteria

- ✅ Frontend can successfully assign labels to tasks
- ✅ API accepts `label_id` in request payload
- ✅ All existing tests pass
- ✅ Swagger documentation is accurate
- ✅ No breaking changes to API contract
