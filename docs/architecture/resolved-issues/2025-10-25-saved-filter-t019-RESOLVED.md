# Saved Filter Bug - Technical Debt

## Problem
Saved filters are not working - they load but display all tasks instead of filtering them.

## Root Cause
The service layer refactor is incomplete. Filter parsing is commented out at line 248 of `pkg/services/task.go`:

```go
// TODO: Implement filter parsing in service layer
// opts.parsedFilters, err = getTaskFiltersFromFilterString(tf.Filter, tf.FilterTimezone)
```

## What's Working
1. ✅ Saved filter loads correctly from database (`filter = "done = false && labels = 5"`)
2. ✅ Filter string merges correctly with incoming URL parameters
3. ✅ Filter string is passed to query options

## What's NOT Working
❌ **The filter is never parsed and applied to the database query**

The filter parsing logic exists in `pkg/models/task_collection_filter.go` but needs to be implemented in the service layer following the new architecture.

## Solution Required

### 1. Move Filter Types to Service Layer
Move from `pkg/models/task_collection_filter.go` to `pkg/services/task.go`:
- `taskFilter` type
- `taskFilterComparator` type and constants
- `taskFilterConcatinator` type and constants

### 2. Implement Filter Parsing
Implement in service layer (currently at line ~248 in `pkg/services/task.go`):
```go
func (ts *TaskService) getTaskFiltersFromFilterString(filter string, filterTimezone string) ([]*taskFilter, error)
```

This needs to include helper functions:
- `parseFilterFromExpression`
- `validateTaskFieldComparator`
- `getFilterComparatorFromOp`
- `parseTimeFromUserInput`
- `validateTaskField`
- `getNativeValueForTaskField`

### 3. Implement Filter to Query Conversion
Implement in service layer:
```go
func (ts *TaskService) convertFiltersToDBFilterCond(filters []*taskFilter, includeNulls bool) (builder.Cond, error)
```

This converts parsed filters into XORM builder conditions that can be applied to the database query.

### 4. Apply Filters in buildTaskQuery
Update `buildTaskQuery` method (around line 2853) to actually apply the parsed filters:
```go
if opts.filter != "" {
    // Parse filters if not already parsed
    if opts.parsedFilters == nil {
        opts.parsedFilters, err = ts.getTaskFiltersFromFilterString(opts.filter, opts.filterTimezone)
        if err != nil {
            return nil, nil, err
        }
    }
    
    // Convert to DB conditions and apply
    filterCond, err := ts.convertFiltersToDBFilterCond(opts.parsedFilters, opts.filterIncludeNulls)
    if err != nil {
        return nil, nil, err
    }
    query = query.And(filterCond)
}
```

### 5. Testing
Test with saved filter having `filter = "done = false && labels = 5"` to ensure:
- Tasks are actually filtered (not showing all tasks)
- Filter expressions parse correctly
- Database query applies the filter conditions

## Estimated Effort
~500+ lines of code to move and adapt from models to service layer.

## Priority
**HIGH** - Blocking core functionality (saved filters)

## Related Files
- `pkg/services/task.go` - Main service file needing updates
- `pkg/models/task_collection_filter.go` - Source of filter parsing logic
- `pkg/models/task_search.go` - Reference for how filters are applied in models layer

## Notes
- This must follow the service layer architecture pattern
- Do NOT create shortcuts by calling back to models layer
- Maintain compatibility with existing filter syntax
- Ensure proper error handling for invalid filter expressions
