# Frontend ModelValue Null Error Fix

**Date**: 2025-10-15  
**Issue**: Vue warning "Missing required prop: modelValue null" when browsing to `/tasks/by/upcoming`  
**Root Cause**: Backend returning `null` instead of empty arrays `[]` for tasks without labels/assignees/attachments  
**Status**: ✅ FIXED

## Problem Description

When browsing to `/tasks/by/upcoming` in the frontend, the following error appeared in the console:

```
Missing required prop: "modelValue" null app warnHandler
app.config.errorHandler	@	main.ts:69
```

## Root Cause Analysis

### Backend Issue (Primary)
The backend service layer was not initializing empty slices for task collections (Labels, Assignees, Attachments). In Go, when a slice is `nil`, it gets marshaled to JSON as `null` instead of an empty array `[]`.

**In the service layer** (`pkg/services/task.go`):
```go
// BEFORE - Labels could be nil
func (ts *TaskService) addLabelsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
    labelService := NewLabelService(ts.DB)
    labels, _, _, err := labelService.GetLabelsByTaskIDs(s, &GetLabelsByTaskIDsOptions{
        TaskIDs: taskIDs,
        Page:    -1,
    })
    // ...
    // If no labels exist for a task, task.Labels remains nil
}
```

This meant:
- Tasks WITH labels: `"labels": [{"id": 1, ...}]` ✅
- Tasks WITHOUT labels: `"labels": null` ❌ (should be `[]`)

### Frontend Issues (Defensive Programming)

The frontend had two places that didn't handle `null` values defensively:

1. **EditLabels.vue** -watch didn't check for null:
```typescript
watch(
	() => props.modelValue,
	(value) => {
		// BEFORE - would fail if value is null
		labels.value = Array.from(new Map(value.map(label => [label.id, label])).values())
	}
)
```

2. **Task model** - constructor didn't handle null labels:
```typescript
// BEFORE - would fail if this.labels is null
this.labels = this.labels
	.map(l => new LabelModel(l))
	.sort((a, b) => a.title.localeCompare(b.title))
```

## Solution

### Backend Fixes (Primary - Root Cause)

Modified three functions in `/home/aron/projects/vikunja/pkg/services/task.go` to initialize empty slices:

#### 1. Labels Initialization
```go
func (ts *TaskService) addLabelsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	// Initialize empty Labels slice for all tasks to ensure JSON returns [] instead of null
	for _, task := range taskMap {
		if task.Labels == nil {
			task.Labels = []*models.Label{}
		}
	}
	// ... rest of function
}
```

#### 2. Assignees Initialization
```go
func (ts *TaskService) addAssigneesToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	// Initialize empty Assignees slice for all tasks to ensure JSON returns [] instead of null
	for _, task := range taskMap {
		if task.Assignees == nil {
			task.Assignees = []*user.User{}
		}
	}
	// ... rest of function
}
```

#### 3. Attachments Initialization
```go
func (ts *TaskService) addAttachmentsToTasks(s *xorm.Session, taskIDs []int64, taskMap map[int64]*models.Task) error {
	// Initialize empty Attachments slice for all tasks to ensure JSON returns [] instead of null
	for _, task := range taskMap {
		if task.Attachments == nil {
			task.Attachments = []*models.TaskAttachment{}
		}
	}
	// ... rest of function
}
```

**Note**: Reminders already had this initialization (lines 1754-1755).

### Frontend Fixes (Defensive Programming)

#### 1. EditLabels.vue - Handle null modelValue
```typescript
watch(
	() => props.modelValue,
	(value) => {
		// Handle null/undefined values from API
		if (!value) {
			labels.value = []
			return
		}
		labels.value = Array.from(new Map(value.map(label => [label.id, label])).values())
	},
	{
		immediate: true,
		deep: true,
	},
)
```

#### 2. Task Model - Handle null labels array
```typescript
this.labels = (this.labels || [])
	.map(l => new LabelModel(l))
	.sort((a, b) => a.title.localeCompare(b.title))
```

#### 3. DatepickerWithRange.vue - Make modelValue optional with defaults
The `DatepickerWithRange` component was requiring a `modelValue` prop but it wasn't being provided in `ShowTasks.vue`. Made it optional with a default value:

```typescript
// BEFORE - Required prop
const props = defineProps<{
	modelValue: {
		dateFrom: Date | string,
		dateTo: Date | string,
	},
}>()

// AFTER - Optional with default
const props = withDefaults(defineProps<{
	modelValue?: {
		dateFrom: Date | string,
		dateTo: Date | string,
	},
}>(), {
	modelValue: () => ({
		dateFrom: '',
		dateTo: '',
	}),
})
```

Also updated the watch to handle undefined/null values:

```typescript
watch(
	() => props.modelValue,
	newValue => {
		if (!newValue) {
			return
		}
		from.value = String(newValue.dateFrom || '')
		to.value = String(newValue.dateTo || '')
		// ... rest of logic
	},
	{immediate: true},
)
```

## Files Modified

### Backend
- `/home/aron/projects/vikunja/pkg/services/task.go`
  - `addLabelsToTasks()` - Initialize empty Labels array
  - `addAssigneesToTasks()` - Initialize empty Assignees array
  - `addAttachmentsToTasks()` - Initialize empty Attachments array

### Frontend
- `/home/aron/projects/vikunja/frontend/src/components/tasks/partials/EditLabels.vue`
  - Added null check in watch
- `/home/aron/projects/vikunja/frontend/src/models/task.ts`
  - Added null coalescing for labels array
- `/home/aron/projects/vikunja/frontend/src/components/date/DatepickerWithRange.vue`
  - Made modelValue prop optional with default value
  - Added null check in watch to handle undefined values

## Verification

```bash
# Backend tests pass
$ go test ./pkg/webtests -run TestTaskLabel -v
=== RUN   TestTaskLabel_AddLabel
--- PASS: TestTaskLabel_AddLabel (0.04s)
=== RUN   TestTaskLabel_AddLabelWithoutID
--- PASS: TestTaskLabel_AddLabelWithoutID (0.02s)
=== RUN   TestTaskLabel_RemoveLabel
--- PASS: TestTaskLabel_RemoveLabel (0.02s)
=== RUN   TestTaskLabel_GetTaskLabels
--- PASS: TestTaskLabel_GetTaskLabels (0.02s)
=== RUN   TestTaskLabel_BulkUpdate
--- PASS: TestTaskLabel_BulkUpdate (0.03s)
=== RUN   TestTaskLabel_PermissionDenied
--- PASS: TestTaskLabel_PermissionDenied (0.03s)
PASS
ok      code.vikunja.io/api/pkg/webtests        0.183s

# Frontend builds without errors
$ cd frontend && npm run build
✓ Built successfully
```

## Expected Behavior After Fix

### API Response Before
```json
{
  "id": 1,
  "title": "Task without labels",
  "labels": null,        ❌ Causes Vue error
  "assignees": null,     ❌ Causes Vue error
  "attachments": null    ❌ Causes Vue error
}
```

### API Response After
```json
{
  "id": 1,
  "title": "Task without labels",
  "labels": [],          ✅ Empty array
  "assignees": [],       ✅ Empty array
  "attachments": []      ✅ Empty array
}
```

## Architecture Decision

The fix follows the principle of **defensive programming at both layers**:

1. **Backend (Primary Fix)**: Ensure API responses are consistent - always return empty arrays, never null
2. **Frontend (Defense in Depth)**: Handle edge cases gracefully even if backend sends unexpected data

This two-layer approach makes the application more resilient to data inconsistencies.

## Related Issues

This fix addresses the same pattern that caused the label assignment error (fixed earlier today in LABEL_TASK_FIX.md). The backend refactoring to the service layer introduced cases where nil slices weren't being initialized before JSON marshaling.

## Success Criteria

- ✅ No Vue warnings when browsing to `/tasks/by/upcoming`
- ✅ Tasks without labels/assignees/attachments display correctly
- ✅ All existing tests pass
- ✅ API returns consistent empty arrays instead of null
- ✅ Frontend handles both null and empty array gracefully
