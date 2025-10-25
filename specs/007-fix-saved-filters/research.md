# Research: Saved Filters Implementation

**Date**: 2025-10-25  
**Feature**: 007-fix-saved-filters  
**Source**: Analysis of `~/projects/vikunja_original_main/pkg/models/`

---

## 1. Filter Conversion Logic Analysis

### Original Implementation
**File**: `~/projects/vikunja_original_main/pkg/models/task_search.go`  
**Function**: `convertFiltersToDBFilterCond` (line 159)

### Decision
The filter conversion logic recursively processes an array of `taskFilter` objects and converts them into XORM `builder.Cond` conditions that can be applied to database queries.

### Logic Flow

```go
func convertFiltersToDBFilterCond(rawFilters []*taskFilter, includeNulls bool) (filterCond builder.Cond, err error)
```

**Step 1: Handle Nested Filters**
- If filter value is `[]*taskFilter` (nested expression from parentheses), recursively call `convertFiltersToDBFilterCond`
- This handles complex boolean logic like `(priority > 2 || labels in [5,6]) && done = false`

**Step 2: Detect Subtable Filters**
- Check if filter field is in `subTableFilters` map (labels, assignees, reminders, parent_project)
- Subtable filters require special EXISTS subquery handling instead of direct column comparisons

**Step 3: Process Subtable Filters**
```go
if ok := subTableFilters[f.field]; ok {
    // Special handling for assignees + like operator (skip it)
    if f.field == "assignees" && f.comparator == taskFilterComparatorLike {
        continue
    }
    
    // Convert strict comparators (=, !=, in, not in) to IN for subtable query
    comparator := f.comparator
    if isStrictComparator(f.comparator) {
        comparator = taskFilterComparatorIn  // Always use IN for value selection
    }
    
    // Build filter condition for the subtable field
    filter := getFilterCond(&taskFilter{
        field: subTableFilterParams.FilterableField,  // e.g., "label_id"
        value: f.value,
        comparator: comparator,
        isNumeric: f.isNumeric,
    }, false)  // includeNulls=false for subquery
    
    // Create EXISTS subquery
    filterSubQuery := subTableFilterParams.ToBaseSubQuery().And(filter)
    // Example: SELECT 1 FROM label_tasks WHERE tasks.id = task_id AND label_id IN (5)
    
    // Negate for != and not in operators
    if f.comparator == taskFilterComparatorNotEquals || f.comparator == taskFilterComparatorNotIn {
        filter = builder.NotExists(filterSubQuery)
    } else {
        filter = builder.Exists(filterSubQuery)
    }
    
    // Add NULL check if requested
    if includeNulls && subTableFilterParams.AllowNullCheck {
        // Include tasks that have NO entries in this subtable
        filter = builder.Or(filter, builder.NotExists(subTableFilterParams.ToBaseSubQuery()))
    }
}
```

**Step 4: Process Regular Field Filters**
```go
// Prefix field with table name
if f.field == taskPropertyBucketID {
    f.field = "task_buckets.`bucket_id`"
} else {
    f.field = "tasks.`" + f.field + "`"
}

// Build condition
filter := getFilterCond(f, includeNulls)
```

**Step 5: Combine Filters with Concatenators**
```go
// Combine filters based on their join/concatenator (AND/OR)
for i, f := range dbFilters {
    if len(dbFilters) > i+1 {
        switch rawFilters[i+1].join {
        case filterConcatOr:
            filterCond = builder.Or(filterCond, f, dbFilters[i+1])
        case filterConcatAnd:
            filterCond = builder.And(filterCond, f, dbFilters[i+1])
        }
    }
}
```

### Rationale
- **EXISTS subqueries prevent duplicate results**: Using JOIN with labels/assignees can produce duplicate task rows; EXISTS returns true/false per task
- **Recursive handling enables complex expressions**: Parentheses in filter syntax create nested filter arrays
- **NULL handling is context-aware**: Regular fields use `OR field IS NULL`, subtables use `OR NOT EXISTS (SELECT 1 FROM subtable...)`
- **Strict comparators normalized to IN**: For subtable queries, = and != are converted to IN and NOT IN for consistency

### Service Layer Adaptation
Port `convertFiltersToDBFilterCond` as a method on `TaskService`:
```go
func (ts *TaskService) convertFiltersToDBFilterCond(rawFilters []*taskFilter, includeNulls bool) (filterCond builder.Cond, err error)
```

Move `subTableFilters` map and `SubTableFilter` type to `pkg/services/task.go` as unexported types. The Chef layer owns all business logic including filter conversion.

---

## 2. Filter Condition Building

### Original Implementation
**File**: `~/projects/vikunja_original_main/pkg/models/tasks.go`  
**Function**: `getFilterCond` (around line 1500)

### Decision
Individual filter conditions are built by mapping `taskFilterComparator` values to XORM builder condition types, with special handling for NULL values when `includeNulls=true`.

### Comparator Mapping

```go
func getFilterCond(f *taskFilter, includeNulls bool) (cond builder.Cond, err error) {
    field := f.field
    
    switch f.comparator {
    case taskFilterComparatorEquals:
        cond = &builder.Eq{field: f.value}
    case taskFilterComparatorNotEquals:
        cond = &builder.Neq{field: f.value}
    case taskFilterComparatorGreater:
        cond = &builder.Gt{field: f.value}
    case taskFilterComparatorGreateEquals:
        cond = &builder.Gte{field: f.value}
    case taskFilterComparatorLess:
        cond = &builder.Lt{field: f.value}
    case taskFilterComparatorLessEquals:
        cond = &builder.Lte{field: f.value}
    case taskFilterComparatorLike:
        val, is := f.value.(string)
        if !is {
            return nil, ErrInvalidTaskFilterValue{Field: field, Value: f.value}
        }
        cond = &builder.Like{field, "%" + val + "%"}  // Wrap value with wildcards
    case taskFilterComparatorIn:
        cond = builder.In(field, f.value)
    case taskFilterComparatorNotIn:
        cond = builder.NotIn(field, f.value)
    case taskFilterComparatorInvalid:
        // Nothing to do
    }
    
    // Add NULL handling
    if includeNulls {
        cond = builder.Or(cond, &builder.IsNull{field})
        if f.isNumeric {
            // For numeric fields, also include 0 values
            cond = builder.Or(cond, &builder.IsNull{field}, &builder.Eq{field: 0})
        }
    }
    
    return
}
```

### Key Behaviors

1. **LIKE operator**: Automatically wraps value with `%` wildcards for substring matching
2. **NULL handling**: When `includeNulls=true`:
   - All fields: Add `OR field IS NULL`
   - Numeric fields: Also add `OR field = 0` (treats 0 as equivalent to NULL)
3. **Type validation**: LIKE operator validates that value is a string
4. **Invalid comparator**: Returns nil condition (no-op)

### Rationale
- **Wildcard wrapping for LIKE**: Matches user expectation that "title like 'report'" finds "Monthly Report"
- **Numeric zero-as-null**: Many numeric fields (priority, percent_done) use 0 as default/unset value
- **OR combination for NULLs**: Ensures NULL values are included when requested without complex nested logic

### Service Layer Adaptation
Port `getFilterCond` as a method on `TaskService`:
```go
func (ts *TaskService) getFilterCond(f *taskFilter, includeNulls bool) (cond builder.Cond, err error)
```

Use service layer error types (`models.ErrInvalidTaskFilterValue` is already accessible from services).

---

## 3. Subtable Filter Patterns

### Original Implementation
**File**: `~/projects/vikunja_original_main/pkg/models/task_search.go`  
**Lines**: 65-100 (SubTableFilter type and map)

### Decision
Subtable filters are defined in a map structure that specifies how to join related tables (labels, assignees, reminders) and query them via EXISTS subqueries.

### SubTableFilter Structure

```go
type SubTableFilter struct {
    Table           string  // Related table name (e.g., "label_tasks")
    BaseFilter      string  // Join condition (e.g., "tasks.id = task_id")
    FilterableField string  // Column to filter on (e.g., "label_id")
    AllowNullCheck  bool    // Whether to support includeNulls for this field
}

func (sf *SubTableFilter) ToBaseSubQuery() *builder.Builder {
    var cond = builder.
        Select("1").
        From(sf.Table).
        Where(builder.Expr(sf.BaseFilter))
    
    // Special case for assignees: also join users table
    if sf.Table == "task_assignees" {
        cond.Join("INNER", "users", "users.id = user_id")
    }
    
    return cond
}
```

### Subtable Filter Definitions

```go
var subTableFilters = map[string]SubTableFilter{
    "labels": {
        Table:           "label_tasks",
        BaseFilter:      "tasks.id = task_id",
        FilterableField: "label_id",
        AllowNullCheck:  true,
    },
    "label_id": {  // Alias for labels
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
        AllowNullCheck:  false,  // No NULL check for parent_project
    },
    "parent_project_id": {  // Alias
        Table:           "projects",
        BaseFilter:      "tasks.project_id = id",
        FilterableField: "parent_project_id",
        AllowNullCheck:  false,
    },
}
```

### Strict Comparators

```go
var strictComparators = map[taskFilterComparator]bool{
    taskFilterComparatorIn:        true,
    taskFilterComparatorNotIn:     true,
    taskFilterComparatorEquals:    true,
    taskFilterComparatorNotEquals: true,
}
```

These comparators are always converted to IN/NOT IN for subtable queries to normalize the query pattern.

### Example Query Generation

**Filter**: `labels = 5`

**Step 1**: Detect it's a subtable filter  
**Step 2**: Convert `=` to `IN` (strict comparator)  
**Step 3**: Build subquery condition:
```sql
SELECT 1 FROM label_tasks 
WHERE tasks.id = task_id 
AND label_id IN (5)
```
**Step 4**: Wrap in EXISTS:
```sql
EXISTS (SELECT 1 FROM label_tasks WHERE tasks.id = task_id AND label_id IN (5))
```

**Filter**: `labels != 5`

**Step 1-3**: Same as above  
**Step 4**: Wrap in NOT EXISTS:
```sql
NOT EXISTS (SELECT 1 FROM label_tasks WHERE tasks.id = task_id AND label_id IN (5))
```

**Filter**: `labels = 5` with `includeNulls=true`

**Step 1-4**: Same as first example  
**Step 5**: Add NULL check (AllowNullCheck=true):
```sql
EXISTS (SELECT 1 FROM label_tasks WHERE tasks.id = task_id AND label_id IN (5))
OR NOT EXISTS (SELECT 1 FROM label_tasks WHERE tasks.id = task_id)
```
This includes tasks that have label 5 OR tasks that have NO labels at all.

### Rationale
- **EXISTS prevents duplicates**: A task with multiple labels would appear multiple times with JOIN
- **SELECT 1 is efficient**: EXISTS only checks existence, doesn't need to return data
- **Strict comparator normalization**: Simplifies subquery logic by always using IN/NOT IN
- **AllowNullCheck flag**: Not all subtable relationships make sense to include "no entries" (e.g., parent_project)

### Service Layer Adaptation
Move `SubTableFilter` type and `subTableFilters` map to `pkg/services/task.go` as unexported:
```go
type subTableFilter struct { ... }  // Lowercase = unexported
var subTableFilters = map[string]subTableFilter{ ... }
```

Include `strictComparators` map as well. All filter infrastructure lives in the Chef layer.

---

## 4. Query Application Strategy

### Original Implementation
**File**: `~/projects/vikunja_original_main/pkg/models/task_search.go`  
**Context**: Filter application in database query building

### Decision
Filters are applied to the XORM query session using the `.And()` method after other conditions (project access, search) but before sorting and pagination.

### Integration Pattern

From analysis of original code:
```go
// Build base query
query := s.Where("tasks.project_id IN (?)", projectIDs)

// Apply search if present
if opts.search != "" {
    query = query.Where("title LIKE ?", "%"+opts.search+"%")
}

// Apply filters if present
if opts.parsedFilters != nil {
    filterCond, err := convertFiltersToDBFilterCond(opts.parsedFilters, opts.filterIncludeNulls)
    if err != nil {
        return nil, 0, err
    }
    query = query.And(filterCond)
}

// Apply sorting
if opts.sortby != nil {
    // ... apply orderBy ...
}

// Apply pagination
if opts.page > 0 {
    query = query.Limit(opts.perPage, (opts.page-1)*opts.perPage)
}
```

### Current Service Layer Location

**File**: `~/projects/vikunja/pkg/services/task.go`  
**Lines**: 3227-3238 (current placeholder)

```go
// Apply custom filters if present
if opts.filter != "" {
    // For now, just delegate back to models for complex filtering
    // This will be moved to service layer in a future iteration
    // For simple cases, we handle here; for complex, we delegate
    if strings.Contains(opts.filter, ">=") || strings.Contains(opts.filter, "<=") ||
        strings.Contains(opts.filter, "!=") || strings.Contains(opts.filter, "&&") ||
        strings.Contains(opts.filter, "||") {
        // Complex filter - delegate to models for now
        // This is where the date range logic and other complex filtering happens
        // TODO: Implement full filter parsing in service layer
    }
}
```

**REPLACE WITH**:
```go
// Apply filters if present
if opts.parsedFilters != nil && len(opts.parsedFilters) > 0 {
    filterCond, err := ts.convertFiltersToDBFilterCond(opts.parsedFilters, opts.filterIncludeNulls)
    if err != nil {
        return nil, nil, err
    }
    query = query.And(filterCond)
}
```

### Order of Operations
1. Base query conditions (project access, bucket filtering if applicable)
2. Search text filtering
3. **Custom filters (ADDED HERE)**
4. Sorting
5. Pagination

### Rationale
- **After search, before sorting**: Filters narrow the result set before sorting for efficiency
- **Use .And() for combining**: Ensures filters are combined with AND logic with other conditions
- **Check parsedFilters not filter string**: Filter string may be present but parsing may have failed
- **Error propagation**: Filter conversion errors should be returned immediately, not silently ignored

### Service Layer Adaptation
The location is already identified in `buildTaskQuery` method. Replace the placeholder comment block (lines 3227-3238) with the actual filter application logic using the ported `convertFiltersToDBFilterCond` method.

---

## 5. Edge Case Handling

### Edge Case 1: Deleted Label/Assignee IDs

**Original Behavior**: 
- Filter contains: `labels = 5`
- Label ID 5 has been deleted from database
- Result: EXISTS subquery returns FALSE for all tasks (no rows in label_tasks with label_id=5)
- **Outcome**: No tasks returned (silently excludes, no error)

**Service Layer Adaptation**: Same behavior - let database naturally handle missing IDs

### Edge Case 2: Malformed Filter Expressions

**Original Behavior**:
- Filter contains: `done = false &&` (incomplete expression)
- fexpr parser returns error
- Result: `ErrInvalidFilterExpression` with parser error details
- **Outcome**: HTTP 400 error to user with explanation

**Service Layer Adaptation**: Same - rely on fexpr library validation, wrap in service error

### Edge Case 3: Invalid Timezone

**Original Behavior**:
- FilterTimezone contains: `"Mars/Olympus"`
- `time.LoadLocation("Mars/Olympus")` returns error
- Result: `ErrInvalidTimezone` error
- **Outcome**: HTTP 400 error to user

**Service Layer Adaptation**: Already implemented in `getTaskFiltersFromFilterString` method

### Edge Case 4: NULL Value Comparisons

**Original Behavior**:
- Filter contains: `due_date = '2025-01-01'`
- Some tasks have `due_date = NULL`
- With `filterIncludeNulls=false`: NULL tasks excluded
- With `filterIncludeNulls=true`: NULL tasks included via `OR due_date IS NULL`

**Service Layer Adaptation**: Implement in `getFilterCond` method (documented in section 2)

### Edge Case 5: Large IN Clauses

**Original Behavior**:
- Filter contains: `labels in [1,2,3,...,1000]`
- No explicit limit on IN clause size
- Database handles it (MySQL/PostgreSQL support large IN clauses)
- **Outcome**: Works but may be slow

**Service Layer Adaptation**: Same - no artificial limit, let database handle it

### Edge Case 6: Date Parsing Ambiguity

**Original Behavior**:
Format precedence in `parseTimeFromUserInput`:
1. RFC3339: `2025-01-01T15:04:05Z`
2. Safari date-time: `2025-01-01 15:04`
3. Safari date: `2025-01-01`
4. Manual parsing: `2025-1-1` (splits on `-`)

**Service Layer Adaptation**: Already ported to `parseTimeFromUserInput` in TaskService

### Edge Case 7: Type Mismatches

**Original Behavior**:
- Filter contains: `priority = 'high'` (string value for int field)
- `strconv.ParseInt("high", 10, 64)` returns error
- Result: `ErrInvalidTaskFilterValue`
- **Outcome**: HTTP 400 error to user

**Service Layer Adaptation**: Already implemented in `getNativeValueForTaskField` method

### Edge Case 8: Invalid Field Names

**Original Behavior**:
- Filter contains: `nonexistent_field = 5`
- `validateTaskField("nonexistent_field")` returns error
- Result: `ErrInvalidTaskField`
- **Outcome**: HTTP 400 error to user

**Service Layer Adaptation**: Already implemented in `validateTaskField` method

### Testing Strategy
Create comprehensive test cases in `pkg/services/task_test.go` for all edge cases:
```go
func TestTaskService_FilterEdgeCases(t *testing.T) {
    // Test deleted label IDs
    // Test malformed expressions
    // Test invalid timezones
    // Test NULL comparisons
    // Test large IN clauses
    // Test date parsing precedence
    // Test type mismatches
    // Test invalid field names
}
```

---

## Summary

### What Needs to Be Ported

1. **Types and Constants** (from `pkg/models/task_search.go`):
   - `SubTableFilter` type → `subTableFilter` (unexported)
   - `subTableFilters` map
   - `strictComparators` map

2. **Methods** (from `pkg/models/task_search.go` and `pkg/models/tasks.go`):
   - `convertFiltersToDBFilterCond` → `TaskService.convertFiltersToDBFilterCond`
   - `getFilterCond` → `TaskService.getFilterCond`
   - `SubTableFilter.ToBaseSubQuery` → `subTableFilter.toBaseSubQuery` (unexported)

3. **Query Application** (in `pkg/services/task.go`):
   - Replace placeholder at lines 3227-3238 with filter application logic

### What Already Exists in Service Layer

✅ Filter parsing: `getTaskFiltersFromFilterString`  
✅ Field validation: `validateTaskField`, `validateTaskFieldComparator`  
✅ Type conversion: `getNativeValueForTaskField`, `getValueForField`  
✅ Date parsing: `parseTimeFromUserInput`  
✅ Filter types: `taskFilter`, `taskFilterComparator`, `taskFilterConcatinator`

### What's Missing (Causes the Bug)

❌ Filter-to-query conversion: `convertFiltersToDBFilterCond`  
❌ Condition building: `getFilterCond`  
❌ Subtable infrastructure: `subTableFilter`, `subTableFilters` map  
❌ Query application: Actual `.And(filterCond)` in buildTaskQuery

### Estimated Effort

- Port 3 types/maps: ~50 lines
- Port 2 methods with adaptations: ~200 lines
- Add query application: ~10 lines
- Comprehensive tests: ~300 lines
- **Total**: ~560 lines of code

### Dependencies Confirmed

- ✅ XORM builder package already imported
- ✅ fexpr library already in use
- ✅ Error types already defined
- ✅ Service registry already set up
- ✅ No new external dependencies required

---

**Research Complete**: All unknowns resolved through original implementation analysis. Ready for Phase 1 design.
