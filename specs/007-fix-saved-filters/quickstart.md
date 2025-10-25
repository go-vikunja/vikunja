# Quickstart Guide: Saved Filters Fix

**Feature**: 007-fix-saved-filters  
**Date**: 2025-10-25  
**For**: Developers implementing the saved filters regression fix

---

## Prerequisites

- Go 1.21+ installed
- Mage build tool installed (`go install github.com/magefile/mage@latest`)
- Access to `~/projects/vikunja_original_main` for reference
- Development environment set up (see AGENTS.md)

---

## Quick Overview

**What**: Fix saved filters regression caused by incomplete service layer refactor  
**Where**: `pkg/services/task.go`  
**How**: Port `convertFiltersToDBFilterCond` from models to services and apply filters in query  
**Testing**: Manual test with user "Aron" (password: test) viewing saved filter "Next Actions" at `/projects/-2`

---

## Implementation Checklist

### Phase 1: Write Tests (TDD - Test First!)

1. **Create test file or extend existing**:
   ```bash
   # Tests go in pkg/services/task_test.go
   # Look for existing filter tests and add new ones
   ```

2. **Write failing tests for filter conversion** (see test plan below)

3. **Run tests to confirm they fail**:
   ```bash
   cd ~/projects/vikunja
   export VIKUNJA_SERVICE_ROOTPATH=$(pwd)
   mage test:feature
   ```

### Phase 2: Port Code from Models to Services

1. **Add subTableFilter type to `pkg/services/task.go`**:
   ```go
   // After existing taskFilter type definition (around line 50)
   type subTableFilter struct {
       Table           string
       BaseFilter      string
       FilterableField string
       AllowNullCheck  bool
   }
   
   func (sf *subTableFilter) toBaseSubQuery() *builder.Builder {
       // Port from ~/projects/vikunja_original_main/pkg/models/task_search.go line ~102
   }
   ```

2. **Add subTableFilters map**:
   ```go
   // After subTableFilter type
   var subTableFilters = map[string]subTableFilter{
       // Port from ~/projects/vikunja_original_main/pkg/models/task_search.go line ~70
   }
   
   var strictComparators = map[taskFilterComparator]bool{
       // Port from ~/projects/vikunja_original_main/pkg/models/task_search.go line ~95
   }
   ```

3. **Add getFilterCond method to TaskService**:
   ```go
   // Add after existing filter parsing methods (around line 500)
   func (ts *TaskService) getFilterCond(f *taskFilter, includeNulls bool) (cond builder.Cond, err error) {
       // Port from ~/projects/vikunja_original_main/pkg/models/tasks.go line ~1500
   }
   ```

4. **Add convertFiltersToDBFilterCond method to TaskService**:
   ```go
   // Add after getFilterCond
   func (ts *TaskService) convertFiltersToDBFilterCond(rawFilters []*taskFilter, includeNulls bool) (filterCond builder.Cond, err error) {
       // Port from ~/projects/vikunja_original_main/pkg/models/task_search.go line ~159
   }
   ```

5. **Apply filters in buildTaskQuery**:
   ```go
   // Find the placeholder comment at line ~3227-3238
   // REPLACE with:
   if opts.parsedFilters != nil && len(opts.parsedFilters) > 0 {
       filterCond, err := ts.convertFiltersToDBFilterCond(opts.parsedFilters, opts.filterIncludeNulls)
       if err != nil {
           return nil, nil, err
       }
       query = query.And(filterCond)
   }
   ```

### Phase 3: Run Tests

1. **Run unit tests**:
   ```bash
   export VIKUNJA_SERVICE_ROOTPATH=$(pwd)
   mage test:feature
   ```

2. **Run web tests** (if applicable):
   ```bash
   mage test:web
   ```

3. **Fix any failing tests** until all pass

### Phase 4: Manual Testing

1. **Start the backend**:
   ```bash
   ./vikunja
   ```

2. **Start the frontend** (in separate terminal):
   ```bash
   cd frontend
   pnpm dev
   ```

3. **Login as test user**:
   - Username: `Aron`
   - Password: `test`

4. **Navigate to saved filter**:
   - URL: `http://localhost:4173/projects/-2`
   - This is the "Next Actions" saved filter

5. **Verify filter works**:
   - ✅ Should show ONLY tasks that are not done AND have @NextActions label
   - ❌ Should NOT show all tasks
   - ✅ Should respect the filter: `"done = false && labels = 5"`

### Phase 5: Lint and Format

1. **Format code**:
   ```bash
   mage fmt
   ```

2. **Fix linting issues**:
   ```bash
   mage lint:fix
   ```

3. **Verify lint passes**:
   ```bash
   mage lint
   ```

### Phase 6: Commit

1. **Review changes**:
   ```bash
   git diff
   ```

2. **Commit with conventional commit message**:
   ```bash
   git add pkg/services/task.go pkg/services/task_test.go
   git commit -m "fix: restore saved filters functionality by implementing filter query application

- Port convertFiltersToDBFilterCond from models to service layer
- Port getFilterCond from models to service layer
- Add subTableFilter infrastructure for labels/assignees/reminders
- Apply parsed filters to database query in buildTaskQuery
- Add comprehensive test coverage for all filter operators
- Maintain 100% feature parity with original implementation

Fixes regression introduced during service layer refactor where filters were
parsed but never applied to database queries, causing all tasks to be returned
instead of filtered results.

Tested with saved filter 'Next Actions' (done = false && labels = 5)"
   ```

---

## Test Plan

### Unit Tests to Add

Add to `pkg/services/task_test.go`:

```go
func TestTaskService_ConvertFiltersToDBFilterCond(t *testing.T) {
    db.LoadAndAssertFixtures(t)
    s := db.NewSession()
    defer s.Close()
    
    ts := NewTaskService(db.GetEngine())
    
    t.Run("simple equality filter", func(t *testing.T) {
        filters := []*taskFilter{{
            field: "tasks.`done`",
            value: false,
            comparator: taskFilterComparatorEquals,
            concatenator: taskFilterConcatAnd,
        }}
        
        cond, err := ts.convertFiltersToDBFilterCond(filters, false)
        assert.NoError(t, err)
        assert.NotNil(t, cond)
        // Verify condition produces correct SQL
    })
    
    t.Run("labels subtable filter", func(t *testing.T) {
        // Test EXISTS subquery for labels
    })
    
    t.Run("complex boolean expression", func(t *testing.T) {
        // Test nested filters with AND/OR
    })
    
    t.Run("NULL handling with includeNulls", func(t *testing.T) {
        // Test NULL inclusion logic
    })
    
    // Add more tests for all operators and edge cases
}

func TestTaskService_GetFilterCond(t *testing.T) {
    ts := NewTaskService(db.GetEngine())
    
    t.Run("all comparator types", func(t *testing.T) {
        // Test =, !=, >, <, >=, <=, like, in, not in
    })
    
    t.Run("NULL handling for regular fields", func(t *testing.T) {
        // Test OR field IS NULL logic
    })
    
    t.Run("NULL handling for numeric fields", func(t *testing.T) {
        // Test OR field = 0 logic
    })
}

func TestTaskService_SavedFilterExecution(t *testing.T) {
    db.LoadAndAssertFixtures(t)
    s := db.NewSession()
    defer s.Close()
    
    ts := NewTaskService(db.GetEngine())
    u := &user.User{ID: 1}
    
    t.Run("filter actually filters tasks", func(t *testing.T) {
        // Create tasks: some matching filter, some not
        // Execute saved filter
        // Verify ONLY matching tasks returned
    })
}
```

### Edge Case Tests

```go
func TestTaskService_FilterEdgeCases(t *testing.T) {
    t.Run("deleted label ID", func(t *testing.T) {
        // Filter: labels = 999 (non-existent)
        // Expected: Empty result, no error
    })
    
    t.Run("malformed filter expression", func(t *testing.T) {
        // Filter: "done = false &&" (incomplete)
        // Expected: ErrInvalidFilterExpression
    })
    
    t.Run("invalid timezone", func(t *testing.T) {
        // FilterTimezone: "Mars/Olympus"
        // Expected: ErrInvalidTimezone
    })
    
    t.Run("large IN clause", func(t *testing.T) {
        // Filter: labels in [1,2,3,...,100]
        // Expected: Works correctly (performance test)
    })
}
```

---

## Debugging Tips

### If filters still don't work:

1. **Check filter parsing**:
   ```go
   // Add debug logging in getTaskFilterOptsFromCollection
   log.Printf("Parsed filters: %+v", opts.parsedFilters)
   ```

2. **Check SQL query**:
   ```go
   // Add before query execution
   sql, args, _ := query.ToSQL()
   log.Printf("Query SQL: %s, Args: %v", sql, args)
   ```

3. **Compare with original**:
   ```bash
   # Check original implementation
   cd ~/projects/vikunja_original_main
   grep -A 20 "convertFiltersToDBFilterCond" pkg/models/task_search.go
   ```

4. **Verify subtable filter configuration**:
   ```go
   // Check subTableFilters map matches original
   // Especially Table, BaseFilter, FilterableField values
   ```

### Common Issues:

- **Filters parsed but not applied**: Check that `query.And(filterCond)` is actually called
- **Subtable filters don't work**: Verify `subTableFilters` map is correctly defined
- **NULL handling wrong**: Check `includeNulls` parameter is passed through correctly
- **Type conversion errors**: Verify field types match between filter values and database columns

---

## Reference Files

**Original Implementation**:
- `~/projects/vikunja_original_main/pkg/models/task_search.go` - Filter conversion logic
- `~/projects/vikunja_original_main/pkg/models/tasks.go` - Filter condition building
- `~/projects/vikunja_original_main/pkg/models/task_collection_filter.go` - Filter parsing

**Current Implementation**:
- `~/projects/vikunja/pkg/services/task.go` - Service layer (add filter conversion here)
- `~/projects/vikunja/pkg/services/task_test.go` - Service layer tests

**Documentation**:
- `~/projects/vikunja/specs/007-fix-saved-filters/research.md` - Detailed research findings
- `~/projects/vikunja/specs/007-fix-saved-filters/data-model.md` - Data model documentation
- `~/projects/vikunja/AGENTS.md` - Development commands and workflows

---

## Time Estimates

- **Phase 1 (Tests)**: 1-2 hours
- **Phase 2 (Code)**: 2-3 hours
- **Phase 3 (Testing)**: 1 hour
- **Phase 4 (Manual)**: 30 minutes
- **Phase 5 (Lint)**: 15 minutes
- **Phase 6 (Commit)**: 15 minutes

**Total**: 5-7 hours

---

## Success Criteria

✅ All unit tests pass (`mage test:feature`)  
✅ Manual test shows filtered results (not all tasks)  
✅ No lint errors (`mage lint`)  
✅ Code formatted (`mage fmt`)  
✅ 100% feature parity with `~/projects/vikunja_original_main`  
✅ All filter operators work (=, !=, >, <, >=, <=, like, in, not in)  
✅ Complex boolean expressions work (AND/OR/parentheses)  
✅ Subtable filters work (labels, assignees, reminders)  
✅ NULL handling works correctly  
✅ Edge cases handled (deleted entities, invalid input, etc.)

---

**Quickstart Complete**: Follow the phases in order for TDD implementation.
