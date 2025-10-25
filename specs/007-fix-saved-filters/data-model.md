# Data Model: Saved Filters Fix

**Date**: 2025-10-25  
**Feature**: 007-fix-saved-filters

---

## Overview

This feature fix does not introduce new data models but works with existing entities. The data model documentation here describes the entities involved in saved filter execution.

---

## Entities

### 1. SavedFilter (Existing - No Changes)

**Location**: `pkg/models/saved_filters.go`

**Purpose**: Stores user-defined filter configurations that can be reused

**Fields**:
```go
type SavedFilter struct {
    ID          int64            `xorm:"bigint autoincr not null unique pk" json:"id"`
    Title       string           `xorm:"varchar(250) not null" json:"title"`
    Description string           `xorm:"longtext null" json:"description"`
    Filters     *TaskCollection  `xorm:"JSON not null" json:"filters"`
    OwnerID     int64            `xorm:"bigint not null INDEX" json:"-"`
    Owner       *user.User       `xorm:"-" json:"owner"`
    Created     time.Time        `xorm:"created not null" json:"created"`
    Updated     time.Time        `xorm:"updated not null" json:"updated"`
}
```

**Key Attributes**:
- `Filters`: Contains `TaskCollection` with filter string, timezone, sort options
- Maps to pseudo-project ID: `SavedFilter.ID` → `Project.ID = -(SavedFilter.ID + 1)`
  - Example: SavedFilter ID 1 → Project ID -2
- `OwnerID`: User who created the filter (enforces ownership permissions)

**Relationships**:
- Belongs to `User` (owner)
- No database foreign key to tasks (filters are applied at query time)

**Validation Rules**:
- Title required, max 250 characters
- Filters.Filter string must be valid filter expression syntax
- Owner must exist and have permission to view filtered tasks

### 2. TaskCollection (Existing - No Changes)

**Location**: `pkg/models/task_collection.go`

**Purpose**: Request object containing filter and query options

**Fields**:
```go
type TaskCollection struct {
    FilterBy            []string `query:"filter_by" json:"filter_by"`
    FilterValue         []string `query:"filter_value" json:"filter_value"`
    FilterComparator    []string `query:"filter_comparator" json:"filter_comparator"`
    FilterConcat        string   `query:"filter_concat" json:"filter_concat"`
    FilterIncludeNulls  bool     `query:"filter_include_nulls" json:"filter_include_nulls"`
    FilterTimezone      string   `query:"filter_timezone" json:"filter_timezone"`
    Filter              string   `query:"filter" json:"filter"`
    
    SortBy              []string `query:"sort_by" json:"sort_by"`
    SortByArr           []string `query:"sort_by[]" json:"-"`
    OrderBy             []string `query:"order_by" json:"order_by"`
    OrderByArr          []string `query:"order_by[]" json:"-"`
    
    ProjectID           int64    `param:"project" json:"-"`
    ProjectViewID       int64    `param:"view" json:"project_view_id"`
    
    // ... other fields for pagination, expand, etc.
}
```

**Key Attributes**:
- `Filter`: Raw filter string (e.g., `"done = false && labels = 5"`)
- `FilterTimezone`: Timezone for date parsing (e.g., `"GMT"`, `"America/New_York"`)
- `FilterIncludeNulls`: Whether to include NULL values in comparisons
- `ProjectID`: Can be negative for saved filters (e.g., `-2` for SavedFilter ID 1)

**Validation Rules**:
- Filter string must parse successfully via fexpr library
- Timezone must be valid `time.Location` name
- Field names in filter must match valid task properties

### 3. Task (Existing - No Changes to Schema)

**Location**: `pkg/models/tasks.go`

**Purpose**: Core entity representing a todo item

**Filterable Fields**:
```go
type Task struct {
    ID              int64      `xorm:"bigint autoincr not null unique pk"`
    Title           string     `xorm:"varchar(250) not null"`
    Description     string     `xorm:"longtext null"`
    Done            bool       `xorm:"INDEX null"`
    DoneAt          time.Time  `xorm:"DATETIME null 'done_at'"`
    DueDate         time.Time  `xorm:"DATETIME null 'due_date'"`
    CreatedByID     int64      `xorm:"bigint not null"`
    ProjectID       int64      `xorm:"bigint INDEX not null"`
    RepeatAfter     int64      `xorm:"bigint INDEX null"`
    Priority        int64      `xorm:"bigint null"`
    StartDate       time.Time  `xorm:"DATETIME null 'start_date'"`
    EndDate         time.Time  `xorm:"DATETIME null 'end_date'"`
    HexColor        string     `xorm:"varchar(6) null"`
    PercentDone     float64    `xorm:"DOUBLE null"`
    UID             string     `xorm:"varchar(250) null"`
    Created         time.Time  `xorm:"created not null"`
    Updated         time.Time  `xorm:"updated not null"`
    Position        float64    `xorm:"double null"`
    BucketID        int64      `xorm:"bigint null"`
    Index           int64      `xorm:"bigint null"`
    
    // Relations (not direct columns)
    Assignees       []*user.User      `xorm:"-" json:"assignees"`
    Labels          []*Label          `xorm:"-" json:"labels"`
    Reminders       []*TaskReminder   `xorm:"-" json:"reminders"`
    // ... other relations
}
```

**Filterable Relationships** (via subtable filters):
- **Labels**: via `label_tasks` join table
- **Assignees**: via `task_assignees` join table
- **Reminders**: via `task_reminders` table
- **Parent Project**: via `projects` table

**Filter Field Aliases**:
- `project` → `project_id`
- `label_id` → uses `labels` subtable filter
- `parent_project` → uses projects subtable filter

---

## Internal Service Types (New - Added to Service Layer)

### 1. taskFilter (Existing in Services - No Changes)

**Location**: `pkg/services/task.go` (already exists)

**Purpose**: Internal representation of a parsed filter expression

```go
type taskFilter struct {
    field        string
    value        interface{}  // Can be primitive or []*taskFilter for nested
    comparator   taskFilterComparator
    concatenator taskFilterConcatinator
    isNumeric    bool
}
```

### 2. subTableFilter (New - To Be Added)

**Location**: `pkg/services/task.go` (to be added)

**Purpose**: Configuration for subtable filter queries (labels, assignees, etc.)

```go
type subTableFilter struct {
    Table           string  // Related table name (e.g., "label_tasks")
    BaseFilter      string  // Join condition (e.g., "tasks.id = task_id")
    FilterableField string  // Column to filter on (e.g., "label_id")
    AllowNullCheck  bool    // Whether to support filterIncludeNulls
}

func (sf *subTableFilter) toBaseSubQuery() *builder.Builder {
    var cond = builder.
        Select("1").
        From(sf.Table).
        Where(builder.Expr(sf.BaseFilter))
    
    // Special case: assignees also join users table
    if sf.Table == "task_assignees" {
        cond.Join("INNER", "users", "users.id = user_id")
    }
    
    return cond
}
```

**Instances**:
```go
var subTableFilters = map[string]subTableFilter{
    "labels": {
        Table:           "label_tasks",
        BaseFilter:      "tasks.id = task_id",
        FilterableField: "label_id",
        AllowNullCheck:  true,
    },
    "label_id": {
        Table:           "label_tasks",
        BaseFilter:      "tasks.id = task_id",
        FilterableField: "label_id",
        AllowNullCheck:  true,
    },
    "reminders": {
        Table:           "task_reminders",
        BaseFilter:      "tasks.id = task_id",
        FilterableField: "reminder",
        AllowNullCheck:  true,
    },
    "assignees": {
        Table:           "task_assignees",
        BaseFilter:      "tasks.id = task_id",
        FilterableField: "username",
        AllowNullCheck:  true,
    },
    "parent_project": {
        Table:           "projects",
        BaseFilter:      "tasks.project_id = id",
        FilterableField: "parent_project_id",
        AllowNullCheck:  false,
    },
    "parent_project_id": {
        Table:           "projects",
        BaseFilter:      "tasks.project_id = id",
        FilterableField: "parent_project_id",
        AllowNullCheck:  false,
    },
}
```

---

## Database Queries

### Query Pattern: Regular Field Filter

**Filter**: `priority >= 3`

**Generated SQL**:
```sql
SELECT * FROM tasks
WHERE tasks.project_id IN (...)
  AND tasks.priority >= 3
ORDER BY ...
LIMIT ...
```

### Query Pattern: Subtable Filter with EXISTS

**Filter**: `labels = 5`

**Generated SQL**:
```sql
SELECT * FROM tasks
WHERE tasks.project_id IN (...)
  AND EXISTS (
    SELECT 1 FROM label_tasks
    WHERE tasks.id = task_id
      AND label_id IN (5)
  )
ORDER BY ...
LIMIT ...
```

### Query Pattern: NOT IN for Subtable

**Filter**: `labels not in [5,6]`

**Generated SQL**:
```sql
SELECT * FROM tasks
WHERE tasks.project_id IN (...)
  AND NOT EXISTS (
    SELECT 1 FROM label_tasks
    WHERE tasks.id = task_id
      AND label_id IN (5, 6)
  )
ORDER BY ...
LIMIT ...
```

### Query Pattern: NULL Inclusion (Regular Field)

**Filter**: `due_date >= '2025-01-01'` with `filterIncludeNulls=true`

**Generated SQL**:
```sql
SELECT * FROM tasks
WHERE tasks.project_id IN (...)
  AND (
    tasks.due_date >= '2025-01-01 00:00:00'
    OR tasks.due_date IS NULL
  )
ORDER BY ...
LIMIT ...
```

### Query Pattern: NULL Inclusion (Subtable Field)

**Filter**: `labels = 5` with `filterIncludeNulls=true`

**Generated SQL**:
```sql
SELECT * FROM tasks
WHERE tasks.project_id IN (...)
  AND (
    EXISTS (
      SELECT 1 FROM label_tasks
      WHERE tasks.id = task_id
        AND label_id IN (5)
    )
    OR NOT EXISTS (
      SELECT 1 FROM label_tasks
      WHERE tasks.id = task_id
    )
  )
ORDER BY ...
LIMIT ...
```

### Query Pattern: Complex Boolean Expression

**Filter**: `(priority > 2 || labels in [5,6]) && done = false`

**Generated SQL**:
```sql
SELECT * FROM tasks
WHERE tasks.project_id IN (...)
  AND (
    (
      tasks.priority > 2
      OR EXISTS (
        SELECT 1 FROM label_tasks
        WHERE tasks.id = task_id
          AND label_id IN (5, 6)
      )
    )
    AND tasks.done = 0
  )
ORDER BY ...
LIMIT ...
```

---

## State Transitions

**Not Applicable**: This is a bug fix, not a feature with state transitions. Filter execution is stateless - each request parses the filter string, converts to query conditions, and executes the query.

---

## Performance Considerations

### Indexes Required (Already Exist)
- `tasks.project_id` - INDEX for project filtering
- `tasks.done` - INDEX for done status filtering
- `label_tasks.task_id` - INDEX for label subqueries
- `label_tasks.label_id` - INDEX for label value filtering
- `task_assignees.task_id` - INDEX for assignee subqueries
- `task_assignees.user_id` - INDEX for assignee value filtering

### Query Optimization Strategies
1. **EXISTS over JOIN**: Prevents duplicate task rows when tasks have multiple labels/assignees
2. **SELECT 1**: Minimal data returned from subqueries (only checks existence)
3. **Indexed subquery joins**: All subtable BaseFilter conditions use indexed columns
4. **Filter before sort**: Apply filters to reduce result set before sorting
5. **Pagination at database level**: LIMIT/OFFSET applied in query, not in-memory

### Expected Performance
- Simple filters (single field): <50ms
- Complex filters (multiple fields + boolean logic): <200ms
- Subtable filters (labels/assignees): <100ms (with proper indexes)
- **No performance regression** expected compared to original implementation (identical query patterns)

---

## Changes Summary

**Schema Changes**: None  
**New Tables**: None  
**New Columns**: None  
**New Indexes**: None  

**Service Layer Additions**:
- `subTableFilter` type
- `subTableFilters` map
- `TaskService.convertFiltersToDBFilterCond` method
- `TaskService.getFilterCond` method
- `subTableFilter.toBaseSubQuery` method

**Models Layer Changes**:
- Deprecate (but keep) `convertFiltersToDBFilterCond` in `pkg/models/task_search.go`
- Deprecate (but keep) `getFilterCond` in `pkg/models/tasks.go`
- Add deprecation comments pointing to service layer

---

**Data Model Complete**: All entities and relationships documented. No schema changes required.
