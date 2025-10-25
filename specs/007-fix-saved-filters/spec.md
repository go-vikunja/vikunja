# Feature Specification: Saved Filters Regression Fix

**Feature Branch**: `007-fix-saved-filters`  
**Created**: 2025-10-25  
**Status**: Draft  
**Priority**: PRIO 1 - Critical Bug Fix
**Input**: User description: "The saved filters feature has stopped working after the service layer refactor. This is PRIO 1 to fix and verify it works 100% like it did in ../vikunja_original_main - while adhering strictly to the new service layer architecture. The backend and frontend need to be assessed and the logic need to be compared to ../vikunja_original_main. If we find a pattern which has broken more functionality, these regressions should be covered and fixed by this spec as well."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Basic Saved Filter Execution (Priority: P1)

A user has previously created a saved filter with criteria such as "show tasks that are not done and have label ID 5". When the user clicks on this saved filter in the sidebar, they expect to see only the tasks matching those criteria, not all tasks.

**Why this priority**: This is the core functionality of saved filters. Without this working, saved filters are completely broken and unusable, making the feature non-functional.

**Independent Test**: Can be fully tested by creating a saved filter with simple criteria (e.g., `done = false && labels = 5`), clicking on it, and verifying that only matching tasks appear in the results.

**Acceptance Scenarios**:

1. **Given** a saved filter exists with criteria "done = false && labels = 5", **When** the user clicks on the saved filter, **Then** only tasks that are not done AND have label 5 are displayed
2. **Given** a saved filter exists with criteria "priority >= 3", **When** the user clicks on the saved filter, **Then** only tasks with priority 3 or higher are displayed
3. **Given** a saved filter exists with date range criteria "due_date >= '2025-01-01' && due_date <= '2025-12-31'", **When** the user clicks on the saved filter, **Then** only tasks with due dates in 2025 are displayed

---

### User Story 2 - Complex Filter Expressions (Priority: P1)

A user creates saved filters with complex boolean logic combining multiple conditions with AND (&&) and OR (||) operators, comparison operators (=, !=, >, <, >=, <=, like), and array operators (in, not in).

**Why this priority**: Complex filters are essential for power users to segment their tasks effectively. This is core to the saved filters value proposition.

**Independent Test**: Can be fully tested by creating a saved filter with complex criteria (e.g., `(priority > 2 || labels in [5,6]) && done = false`), executing it, and verifying results match the logical expression.

**Acceptance Scenarios**:

1. **Given** a saved filter with criteria "(priority > 2 || labels in [5,6]) && done = false", **When** the user executes the filter, **Then** tasks shown must either have priority > 2 OR have labels 5 or 6, AND must not be done
2. **Given** a saved filter with "title like '%report%' && assignees = 1", **When** the user executes the filter, **Then** only tasks with "report" in the title assigned to user ID 1 are shown
3. **Given** a saved filter with "due_date <= 'now' && done != true", **When** the user executes the filter, **Then** only overdue incomplete tasks are displayed

---

### User Story 3 - Date and Time Filtering (Priority: P1)

A user creates saved filters using date-based criteria with various date formats (RFC3339, Safari date format, simple date format) and relative date expressions (e.g., "now", "now+7d").

**Why this priority**: Date filtering is one of the most common use cases for saved filters (overdue tasks, upcoming tasks, tasks due this week, etc.).

**Independent Test**: Can be fully tested by creating saved filters with date criteria using different formats and relative expressions, then verifying the correct date filtering is applied.

**Acceptance Scenarios**:

1. **Given** a saved filter with "due_date >= 'now'", **When** the user executes the filter, **Then** only tasks due today or in the future are shown
2. **Given** a saved filter with date in Safari format "2025-10-25 14:30", **When** the user executes the filter, **Then** the date is correctly parsed and filter applied
3. **Given** a saved filter with timezone-specific date criteria, **When** the user executes the filter in their timezone, **Then** dates are correctly interpreted in the specified timezone

---

### User Story 4 - Filter Field Validation (Priority: P2)

A user attempts to create or use a saved filter with invalid field names, operators, or value types. The system provides clear error messages.

**Why this priority**: Error handling ensures users understand when they've made a mistake and can correct it. While important, the system can function without perfect error messages.

**Independent Test**: Can be fully tested by attempting to create saved filters with various invalid inputs and verifying appropriate error messages are returned.

**Acceptance Scenarios**:

1. **Given** a saved filter with invalid field "nonexistent_field = 5", **When** the filter is executed, **Then** an error message indicates the field is invalid
2. **Given** a saved filter with invalid comparator "priority << 5", **When** the filter is executed, **Then** an error message indicates the operator is invalid
3. **Given** a saved filter with type mismatch "priority = 'high'", **When** the filter is executed, **Then** an error message indicates the value type is incompatible with the field

---

### User Story 5 - Special Field Handling (Priority: P2)

A user creates saved filters using special fields that require join queries (assignees, labels, reminders) or alias fields (project/project_id).

**Why this priority**: These fields are commonly filtered but require special handling. Without this, many practical filters won't work, but basic filtering still functions.

**Independent Test**: Can be fully tested by creating saved filters using assignees, labels, and reminders fields and verifying the join queries execute correctly.

**Acceptance Scenarios**:

1. **Given** a saved filter with "assignees = 1", **When** the filter is executed, **Then** tasks assigned to user ID 1 are shown (requires join to task_assignees table)
2. **Given** a saved filter with "labels in [5,6,7]", **When** the filter is executed, **Then** tasks with any of those labels are shown (requires join to label_tasks table)
3. **Given** a saved filter using "project" field, **When** the filter is executed, **Then** it is correctly interpreted as "project_id"

---

### User Story 6 - Regression Pattern Detection (Priority: P3)

During the fix, any architectural patterns that broke saved filters are identified and tested to ensure they haven't broken other features using similar patterns.

**Why this priority**: While important for system health, this is about preventing future issues rather than fixing the immediate saved filters problem.

**Independent Test**: Can be independently verified by running comprehensive test suites for features that use similar filter/query patterns.

**Acceptance Scenarios**:

1. **Given** the saved filter fix is implemented, **When** running tests for task search functionality, **Then** all existing task search tests pass
2. **Given** the saved filter fix is implemented, **When** running tests for project views (Kanban, Gantt, Table), **Then** all view filtering continues to work
3. **Given** filter parsing logic is moved to service layer, **When** comparing behavior with original implementation, **Then** 100% functional parity is maintained

---

### Edge Cases

- What happens when a saved filter references a label or assignee that has been deleted? → Database EXISTS subqueries naturally return no matches (same as original implementation)
- How does the system handle saved filters with malformed filter expressions (unbalanced parentheses, missing operators)?
- What happens when filter timezone is invalid or not recognized?
- How are NULL values handled in comparisons (e.g., "due_date = null" vs "due_date != null")? → Controlled by filterIncludeNulls: when true, adds OR conditions for NULL values (and 0 for numeric fields); for subtables, includes tasks with no related entries
- What happens when a saved filter uses a field that exists but shouldn't be filterable (like "created" which is auto-generated)?
- How does the system handle filters with very large IN clauses (e.g., "labels in [1,2,3,...,1000]")?
- What happens when date parsing fails for an ambiguous date format?
- How are numeric vs string comparisons handled for mixed-type fields?

## Requirements *(mandatory)*

### Functional Requirements

**Core Filter Execution:**
- **FR-001**: System MUST parse saved filter criteria from the filter string stored in the database
- **FR-002**: System MUST convert parsed filter criteria into database query conditions
- **FR-003**: System MUST apply filter conditions to task queries before returning results
- **FR-004**: System MUST support all comparison operators: =, !=, >, <, >=, <=, like, in, not in
- **FR-005**: System MUST support boolean concatenators: && (AND), || (OR)
- **FR-006**: System MUST support nested filter expressions with parentheses for complex logic

**Field Handling:**
- **FR-007**: System MUST validate that filter field names correspond to valid task properties
- **FR-008**: System MUST convert filter field values to their native types (int64, string, bool, time.Time, float64)
- **FR-009**: System MUST handle special fields requiring joins: assignees, labels, reminders (all referenced by numeric ID only)
- **FR-010**: System MUST support field aliases (e.g., "project" as alias for "project_id")
- **FR-011**: System MUST handle NULL value comparisons with filterIncludeNulls setting: when true, add "OR field IS NULL" for regular fields (plus "OR field = 0" for numeric fields), and "OR NOT EXISTS (subquery)" for subtable fields with AllowNullCheck=true

**Date and Time Processing:**
- **FR-012**: System MUST parse dates in multiple formats: RFC3339, Safari date-time, Safari date, simple date (YYYY-MM-DD)
- **FR-013**: System MUST support relative date expressions using datemath library (e.g., "now", "now+7d")
- **FR-014**: System MUST apply timezone settings from filter configuration when parsing dates
- **FR-015**: System MUST convert all parsed dates to the configured application timezone

**Error Handling:**
- **FR-016**: System MUST return descriptive errors for invalid field names
- **FR-017**: System MUST return descriptive errors for invalid comparator operators
- **FR-018**: System MUST return descriptive errors for type mismatches between field and value
- **FR-019**: System MUST return descriptive errors for malformed filter expressions
- **FR-020**: System MUST return descriptive errors for invalid timezone names

**Service Layer Architecture:**
- **FR-021**: All filter parsing logic MUST reside in the service layer (pkg/services/task.go)
- **FR-022**: Service layer MUST NOT call back to models layer for filter parsing or conversion
- **FR-023**: Filter type definitions (taskFilter, taskFilterComparator, taskFilterConcatinator) MUST be defined in service layer
- **FR-024**: Database query building MUST occur in service layer using parsed filter objects
- **FR-025**: Saved filter service MUST integrate with task service for filter execution

**Backward Compatibility:**
- **FR-026**: System MUST maintain 100% functional parity with pre-refactor saved filter behavior
- **FR-027**: Existing saved filters in database MUST continue to work without modification
- **FR-028**: Frontend MUST NOT require changes to work with refactored backend

### Key Entities

- **SavedFilter**: Represents a user's saved filter configuration with filter criteria, title, description, and ownership
  - Contains filter string (e.g., "done = false && labels = 5")
  - Contains timezone setting for date interpretation
  - Contains filterIncludeNulls setting for NULL handling
  - Maps to pseudo-project ID (negative project ID) for API access

- **TaskFilter**: Internal representation of parsed filter criteria (service layer)
  - field: The task property being filtered (e.g., "priority", "due_date")
  - value: The native-typed value or nested filter array
  - comparator: The comparison operator (=, !=, >, <, etc.)
  - concatenator: How this filter joins with next (AND/OR)
  - isNumeric: Flag for numeric field optimization

- **TaskCollection**: Request object for task queries
  - Filter: Raw filter string
  - FilterTimezone: Timezone for date parsing
  - FilterIncludeNulls: Whether to include NULL values in comparisons
  - SortBy/OrderBy: Sorting configuration
  - ProjectID: Can be negative for saved filters

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can execute saved filters and see only matching tasks (not all tasks)
- **SC-002**: Filter expressions with all operators (=, !=, >, <, >=, <=, like, in, not in) produce correct results
- **SC-003**: Complex boolean expressions with AND/OR/parentheses filter tasks correctly
- **SC-004**: Date-based filters parse and apply correctly for all supported date formats
- **SC-005**: Saved filters execute with performance comparable to pre-refactor implementation (no significant slowdown)
- **SC-006**: All existing automated tests for saved filters pass
- **SC-007**: Manual testing against ../vikunja_original_main shows 100% functional parity
- **SC-008**: No regression in related features (task search, project views, favorites)

## Clarifications

### Session 2025-10-25

- Q: How should the system handle label references in filter strings (by ID vs by name)? → A: Labels are referenced by numeric ID only; text names require frontend resolution to IDs before sending filter string (matches original implementation behavior for 100% feature parity)
- Q: What should happen when a saved filter references deleted labels or assignees? → A: Execute filter normally; database EXISTS subqueries will naturally return no matches for deleted entity IDs (same as original implementation)
- Q: How should NULL value comparisons work with filterIncludeNulls? → A: When true, for regular fields add "OR field IS NULL" (plus "OR field = 0" for numeric fields); for subtable fields with AllowNullCheck=true, add "OR NOT EXISTS (subquery)" to include tasks with no related entries

## Assumptions

1. **Architecture Decision**: The service layer refactor is the correct long-term architecture, so the fix must conform to it rather than reverting to the old approach
2. **Database Schema**: The saved filters database schema has not changed and contains valid filter strings
3. **Frontend Unchanged**: The frontend is sending the same API requests as before the refactor
4. **Test Coverage**: The original codebase at ../vikunja_original_main represents the correct expected behavior
5. **Performance**: Filter parsing and execution performance similar to the original models-layer implementation is acceptable
6. **Error Handling**: Error messages can be improved during the fix but must not be worse than the original implementation
7. **Timezone Handling**: The existing timezone configuration and parsing logic is correct and should be preserved
8. **NULL Handling**: The filterIncludeNulls behavior from the original implementation is correct and should be maintained

## Dependencies

1. **Original Implementation Access**: Requires access to ../vikunja_original_main for behavior comparison
2. **Service Layer Architecture**: Depends on the established service layer pattern (Chef/Waiter/Pantry)
3. **Existing Libraries**: Uses github.com/ganigeorgiev/fexpr for filter expression parsing (same as original)
4. **Existing Libraries**: Uses github.com/jszwedko/go-datemath for relative date parsing (same as original)
5. **SavedFilter Service**: Requires SavedFilterService to be properly initialized and wired
6. **Task Service**: All filter logic must integrate with TaskService.GetAllWithFullFiltering

## Technical Context

### Root Cause Analysis

The saved filter regression occurred during the service layer refactor when filter parsing logic was moved from `pkg/models/task_collection_filter.go` to `pkg/services/task.go`. While the parsing functions were moved and adapted, the critical step of **applying** the parsed filters to the database query was never implemented in the service layer.

**What was moved:**
- Filter type definitions (taskFilter, taskFilterComparator, taskFilterConcatinator)
- Filter parsing functions (getTaskFiltersFromFilterString, parseFilterFromExpression)
- Field validation and type conversion functions

**What was NOT moved:**
- `ConvertFiltersToDBFilterCond` function that converts parsed filters to XORM builder conditions
- The actual application of filter conditions in the query building logic

**Current State:**
- Filters are parsed into `opts.parsedFilters` successfully
- Filter string is available in `opts.filter`
- BUT the buildTaskQuery method never applies these filters to the database query
- Line 3227-3238 in pkg/services/task.go contains only a placeholder comment for complex filter handling

### Files Requiring Changes

**Service Layer (Backend):**
1. `pkg/services/task.go` - Main implementation file
   - Implement `convertFiltersToDBFilterCond` method
   - Update query building to apply parsed filters
   - Ensure all filter helper functions are complete

2. `pkg/services/saved_filter.go` - May need integration updates
   - Verify proper interaction with TaskService

**Testing:**
3. `pkg/services/task_test.go` - Service layer tests
   - Add comprehensive filter execution tests
   - Add edge case tests
   
4. `pkg/services/saved_filter_test.go` - Integration tests
   - Test saved filter execution end-to-end

**Models Layer (Bridge):**
5. `pkg/models/task_collection_filter.go` - May need cleanup
   - Remove or deprecate functions now in service layer
   - Ensure no duplicate logic

6. `pkg/models/task_search.go` - May need cleanup
   - Remove `ConvertFiltersToDBFilterCond` if fully moved to service layer
   - Verify no orphaned filter code

**Frontend Assessment:**
7. Frontend files (TBD based on comparison with original)
   - Verify API request format hasn't changed
   - Ensure error handling aligns with new error types

### Reference Implementation

The original working implementation in `../vikunja_original_main` contains:
- `pkg/models/task_collection_filter.go` - Complete filter parsing logic
- `pkg/models/task_search.go` - `ConvertFiltersToDBFilterCond` function (line 159)
- `pkg/models/task_search.go` - Filter application in query building (line 274)

These must be ported to the service layer following the Chef/Waiter/Pantry pattern.

