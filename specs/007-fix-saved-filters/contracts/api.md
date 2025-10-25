# API Contracts: Saved Filters Fix

**Date**: 2025-10-25  
**Feature**: 007-fix-saved-filters

---

## Overview

**No API contract changes required.** This is a bug fix that restores existing functionality without modifying the API interface. The API endpoints, request/response formats, and error codes remain unchanged.

---

## Affected Endpoints

### GET `/api/v1/projects/{projectId}/views/{viewId}/tasks`

**Purpose**: Get tasks for a project view (including saved filters)

**Special Behavior**: When `projectId` is negative (e.g., `-2`), it represents a saved filter:
- SavedFilter ID = `-(projectId + 1)`
- Example: `projectId = -2` → SavedFilter ID = 1

**Request**:
```http
GET /api/v1/projects/-2/views/21/tasks?sort_by[]=position&order_by[]=asc&filter_include_nulls=false&filter_timezone=GMT&s=&expand=subtasks&page=1
Authorization: Bearer <jwt_token>
```

**Query Parameters** (Existing - No Changes):
- `sort_by[]`: Array of fields to sort by (e.g., `position`, `priority`, `due_date`)
- `order_by[]`: Array of sort orders (`asc` or `desc`)
- `filter_include_nulls`: Boolean - whether to include NULL values in filter comparisons
- `filter_timezone`: Timezone for date parsing (e.g., `GMT`, `America/New_York`)
- `s`: Search string for title filtering
- `expand`: Comma-separated list of relations to include (e.g., `subtasks`, `assignees`, `labels`)
- `page`: Page number for pagination

**Response** (Existing - No Changes):
```json
[
  {
    "id": 123,
    "title": "Task title",
    "description": "Task description",
    "done": false,
    "done_at": null,
    "due_date": "2025-01-15T00:00:00Z",
    "priority": 3,
    "labels": [
      {
        "id": 5,
        "title": "@NextActions",
        "hex_color": "ff0000"
      }
    ],
    "assignees": [],
    "project_id": 1,
    "created": "2025-01-01T12:00:00Z",
    "updated": "2025-01-10T15:30:00Z"
  }
]
```

**Headers**:
- `X-Pagination-Total-Pages`: Total number of pages
- `X-Pagination-Result-Count`: Number of results on this page
- `X-Total-Count`: Total number of matching tasks

**Bug Being Fixed**:
- **Before Fix**: Returns ALL tasks regardless of saved filter criteria
- **After Fix**: Returns ONLY tasks matching the saved filter criteria

**Error Responses** (Existing - No Changes):
- `400 Bad Request`: Invalid filter expression, invalid timezone, invalid field name
- `401 Unauthorized`: Missing or invalid JWT token
- `403 Forbidden`: User doesn't have permission to view saved filter
- `404 Not Found`: Saved filter doesn't exist
- `500 Internal Server Error`: Database error or unexpected server error

---

## SavedFilter Structure (Reference Only)

The saved filter entity stores the filter configuration:

```json
{
  "id": 1,
  "title": "Next Actions",
  "description": "Tasks that are not done and have @NextActions label",
  "filters": {
    "filter": "done = false && labels = 5",
    "filter_timezone": "GMT",
    "filter_include_nulls": false,
    "sort_by": ["position"],
    "order_by": ["asc"]
  },
  "owner": {
    "id": 1,
    "username": "Aron"
  },
  "created": "2025-01-01T00:00:00Z",
  "updated": "2025-01-10T00:00:00Z"
}
```

**Note**: The `filters.filter` string is the key part being fixed - it must be parsed and applied to the database query.

---

## Filter Expression Syntax (Reference Only)

**No changes to filter syntax.** This documents the existing syntax that must be supported:

### Comparison Operators
- `=` - Equals
- `!=` - Not equals
- `>` - Greater than
- `<` - Less than
- `>=` - Greater than or equal
- `<=` - Less than or equal
- `like` - String contains (case-sensitive substring match)
- `in` - Value in list
- `not in` - Value not in list

### Boolean Operators
- `&&` - AND
- `||` - OR
- `()` - Grouping (parentheses)

### Examples

**Simple Filter**:
```
done = false
```

**Multiple Conditions (AND)**:
```
done = false && priority >= 3
```

**Multiple Conditions (OR)**:
```
priority > 4 || due_date < '2025-01-31'
```

**Complex Boolean Logic**:
```
(priority > 2 || labels in [5,6]) && done = false
```

**Date Filters**:
```
due_date >= '2025-01-01' && due_date <= '2025-12-31'
```

**String Matching**:
```
title like 'report'
```

**Subtable Filters**:
```
labels = 5
assignees = 1
labels in [5,6,7]
```

### Field Reference Clarification

**Labels/Assignees/Reminders**:
- MUST use numeric IDs, not names
- Example: `labels = 5` (correct) NOT `labels = @NextActions` (incorrect)
- Frontend must resolve label names to IDs before sending filter string

**Date Formats Supported**:
- RFC3339: `2025-01-01T15:04:05Z`
- Safari date-time: `2025-01-01 15:04`
- Safari date: `2025-01-01`
- Simple date: `2025-1-1` (manual parsing)
- Relative dates: `now`, `now+7d`, `now-1w` (via datemath library)

---

## Internal Query Flow (Implementation Detail)

This documents how the filter is processed internally (for testing purposes):

```
1. Frontend Request
   ↓
2. Router (pkg/routes/api/v1/)
   Extracts projectId = -2 from URL
   ↓
3. Handler calls TaskService.GetAllWithFullFiltering()
   ↓
4. Service detects negative projectId
   Calls SavedFilterService.GetByIDSimple(filterID=1)
   ↓
5. Service merges saved filter settings with request params
   filter = "done = false && labels = 5"
   filterTimezone = "GMT"
   filterIncludeNulls = false
   ↓
6. Service parses filter string
   TaskService.getTaskFiltersFromFilterString()
   → Returns []*taskFilter
   ↓
7. Service converts to database conditions (FIXED HERE)
   TaskService.convertFiltersToDBFilterCond()
   → Returns builder.Cond
   ↓
8. Service applies conditions to query
   query.And(filterCond)
   ↓
9. Execute query and return results
   ↓
10. Response sent to frontend
```

**The bug**: Step 7 was never implemented, so filters were parsed but never applied to the query.

---

## Testing Contract Compliance

### Test Case 1: Simple Filter
**Request**: `GET /api/v1/projects/-2/views/21/tasks` (SavedFilter: "done = false")  
**Expected**: Returns only incomplete tasks  
**Verify**: All returned tasks have `done = false`

### Test Case 2: Label Filter
**Request**: `GET /api/v1/projects/-2/views/21/tasks` (SavedFilter: "labels = 5")  
**Expected**: Returns only tasks with label ID 5  
**Verify**: All returned tasks have label 5 in `labels` array

### Test Case 3: Complex Filter
**Request**: `GET /api/v1/projects/-2/views/21/tasks` (SavedFilter: "done = false && labels = 5")  
**Expected**: Returns only incomplete tasks with label ID 5  
**Verify**: All returned tasks have `done = false` AND label 5

### Test Case 4: Invalid Filter
**Request**: `GET /api/v1/projects/-2/views/21/tasks` (SavedFilter: "nonexistent_field = 5")  
**Expected**: 400 Bad Request with error message  
**Verify**: Response contains `ErrInvalidTaskField` error

### Test Case 5: Deleted Label
**Request**: `GET /api/v1/projects/-2/views/21/tasks` (SavedFilter: "labels = 999")  
**Expected**: Returns empty array (no tasks have non-existent label)  
**Verify**: Response is `[]` with `X-Total-Count: 0`

### Test Case 6: NULL Handling
**Request**: `GET /api/v1/projects/-2/views/21/tasks?filter_include_nulls=true` (SavedFilter: "due_date >= '2025-01-01'")  
**Expected**: Returns tasks with due_date >= 2025-01-01 OR due_date IS NULL  
**Verify**: Response includes tasks with NULL due dates

---

## API Documentation References

**OpenAPI/Swagger**: No changes required to existing API documentation at `pkg/swagger/`

**Frontend API Client**: No changes required to `frontend/src/services/task.ts`

---

**Contracts Complete**: No API changes. Fix is transparent to API consumers.
