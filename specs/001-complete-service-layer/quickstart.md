# Quickstart: Complete Service-Layer Refactor Stabilization and Validation

## Prerequisites

1. **Environment Setup**
   ```bash
   cd /path/to/vikunja
   devenv shell
   export VIKUNJA_SERVICE_ROOTPATH=$(pwd)
   ```

2. **Verify Current State**
   ```bash
   # Check failing tests
   mage test:feature
   
   # Should show failures in task-related tests
   # Missing: RelatedTasks, Labels, Attachments, Assignees
   ```

3. **Reference System**
   ```bash
   # Ensure original main branch is available
   ls ../vikunja_original_main/
   
   # Should contain working reference implementation
   ```

## Phase 1: System Stabilization (Target: 2-3 days)

### Step 1: Fix Task Query Data Population

1. **Identify Root Cause**
   ```bash
   # Run specific failing test
   go test ./pkg/services -v -run TestTaskCollection_ReadAll
   
   # Expected: Failures showing missing related data
   ```

2. **Fix Service Layer Query Methods**
   ```bash
   # Edit task service implementation
   vim pkg/services/task.go
   
   # Add proper data population for:
   # - RelatedTasks (subtask, parenttask, related)
   # - Labels 
   # - Attachments
   # - Assignees
   ```

3. **Verify Fix**
   ```bash
   go test ./pkg/services -v -run TestTaskCollection_ReadAll
   # Expected: Tests pass with complete data
   ```

### Step 2: Fix Label Creation 404 Error

1. **Identify Handler Issue**
   ```bash
   # Check label handler
   vim pkg/routes/api/v1/label.go
   
   # Verify declarative routing pattern
   # Ensure handler wrapper usage
   ```

2. **Fix Service Integration**
   ```bash
   # Verify label service exists and is called
   vim pkg/services/label.go
   
   # Ensure complete business logic implementation
   ```

3. **Test Label Creation**
   ```bash
   # Start development server
   ./vikunja web
   
   # Test label creation via API or UI
   # Expected: 201 Created, no 404 errors
   ```

### Step 3: Fix Empty Task Detail View

1. **Identify UI Integration Issue**
   ```bash
   # Check task detail API response
   curl -X GET http://localhost:3456/api/v1/tasks/1
   
   # Verify complete data structure returned
   ```

2. **Fix Frontend Integration**
   ```bash
   cd frontend
   # Check task detail component integration
   # Ensure API calls match refactored endpoints
   ```

3. **Verify UI Functionality**
   ```bash
   # Build and test frontend
   cd frontend && pnpm install && pnpm run build && cd ..
   ./vikunja web
   
   # Navigate to task detail view
   # Expected: Complete task information displayed
   ```

### Step 4: Achieve 100% Test Pass Rate

```bash
# Run full test suite
mage test:feature

# Expected output:
# PASS: All tests passing
# No failures in any test suite
```

## Phase 2: Complete Refactor (Target: 4-5 days)

### Step 1: Prioritize Refactor Queue

1. **Dependency Analysis**
   ```bash
   # Review refactoring_analysis.md
   cat refactoring_analysis.md
   
   # Identify dependency chains:
   # Labels → Label-Task Management
   # Projects → Project Views, Permissions
   # Tasks → Bulk Update, Kanban Buckets
   ```

2. **Create Refactor Order (Dependency → Complexity)**
   ```
   Low Complexity First:
   1. Favorites (no dependencies)
   2. User Mentions (depends on Users - already done)
   
   Medium Complexity:
   3. Labels (foundation for Label-Task)
   4. API Tokens (depends on Users)
   5. Reactions (minimal dependencies)
   6. Notifications (depends on Users)
   
   High Complexity (Dependency Order):
   7. Projects (foundation for many features)
   8. Project-User Permissions (depends on Projects)
   9. Project-Team Permissions (depends on Projects)  
   10. Project Views (depends on Projects)
   11. Tasks (foundation for task-related features)
   12. Label-Task Management (depends on Labels + Tasks)
   13. Kanban Buckets (depends on Tasks)
   14. Bulk Task Update (depends on Tasks)
   15. Saved Filters (depends on Projects + Tasks)
   16. Link Sharing (depends on Projects)
   17. Subscriptions (depends on Projects + Tasks)
   18. Duplicate Project (depends on Projects)
   19. User Data Export (depends on all features)
   ```

### Step 2: Service Layer Refactor Template

For each feature, follow this TDD process:

1. **Create Service Tests**
   ```bash
   # Create comprehensive test suite
   touch pkg/services/{feature}_test.go
   
   # Write tests that FAIL initially
   # Cover all business logic scenarios
   ```

2. **Implement Service Layer**
   ```bash
   # Create service with business logic
   touch pkg/services/{feature}.go
   
   # Move ALL business logic from models
   # Follow "Chef, Waiter, Pantry" pattern
   ```

3. **Setup Dependency Inversion**
   ```bash
   # In model file: create function variable
   # In service file: populate in init()
   # Maintain backward compatibility
   ```

4. **Update Handler Layer**
   ```bash
   # Use declarative routing pattern
   # Implement handler wrappers
   # Call service layer instead of models
   ```

5. **Verify Tests Pass**
   ```bash
   go test ./pkg/services -v -run Test{Feature}
   # All new tests must pass
   ```

### Step 3: Validate Each Feature

```bash
# After each feature refactor:
mage test:feature  # Must still pass
go test ./pkg/services -v  # New tests pass
```

## Phase 2.4: Route Modernization (Target: 2-3 days)

### Overview
Migrate all legacy WebHandler routes to modern declarative APIRoute pattern for complete architectural consistency.

### Step 1: Identify Legacy Routes

1. **Audit routes.go**
   ```bash
   # Find all WebHandler declarations
   grep -A 5 "WebHandler" pkg/routes/routes.go
   
   # Current legacy routes (15 handlers):
   # - Task Positions, Bulk Tasks, Task Assignees, Bulk Assignees
   # - Label-Task Associations, Bulk Label-Tasks
   # - Task Relations
   # - Project Teams, Project Users
   # - Subscriptions, Notifications
   # - API Tokens, Webhooks, Teams, Team Members
   ```

2. **Categorize by Feature Group**
   ```bash
   # Group related routes for migration:
   # - task_assignee.go (assignees + bulk)
   # - label_task.go (label associations + bulk)
   # - task_relation.go (task relations)
   # - project_permission.go (teams + users)
   # - subscription.go (subscriptions)
   # - api_token.go (API tokens)
   # - team.go (teams + members)
   ```

### Step 2: Create Modern Route Files

1. **Create API v1 Route File Template**
   ```bash
   # Example: pkg/routes/api/v1/label_task.go
   vim pkg/routes/api/v1/label_task.go
   ```

2. **Implement Declarative Pattern**
   ```go
   package v1
   
   var LabelTaskRoutes = []APIRoute{
       {Method: "PUT", Path: "/tasks/:task/labels", 
        Handler: handler.WithDBAndUser(createLabelTaskLogic, true), 
        PermissionScope: "create"},
       {Method: "DELETE", Path: "/tasks/:task/labels/:label", 
        Handler: handler.WithDBAndUser(deleteLabelTaskLogic, true), 
        PermissionScope: "delete"},
       {Method: "GET", Path: "/tasks/:task/labels", 
        Handler: handler.WithDBAndUser(getAllLabelTasksLogic, false), 
        PermissionScope: "read_all"},
   }
   
   func RegisterLabelTasks(a *echo.Group) {
       registerRoutes(a, LabelTaskRoutes)
   }
   
   func createLabelTaskLogic(s *xorm.Session, u *user.User, c echo.Context) error {
       // Service layer implementation
       lt := new(models.LabelTask)
       if err := c.Bind(lt); err != nil {
           return echo.NewHTTPError(http.StatusBadRequest, err)
       }
       
       // Call service layer methods
       // Return formatted response
   }
   ```

3. **Update routes.go**
   ```bash
   # Replace WebHandler with Register* call
   vim pkg/routes/routes.go
   
   # BEFORE:
   # labelTaskHandler := &handler.WebHandler{...}
   # a.PUT("/tasks/:task/labels", labelTaskHandler.CreateWeb)
   
   # AFTER:
   # apiv1.RegisterLabelTasks(a)
   ```

### Step 3: Migration Sequence

**Priority Order** (migrate in this sequence):
1. **Task-Related Routes** (positions, assignees, relations)
2. **Label-Task Routes** (associations, bulk operations)
3. **Project Permission Routes** (teams, users)
4. **Subscription Routes** (entity subscriptions)
5. **Notification Routes** (bulk operations)
6. **Team Management Routes** (teams, members)
7. **API Token Routes** (token management)
8. **Remaining Routes** (webhooks, etc.)

### Step 4: Test Migration

```bash
# After each route file migration:
go test ./pkg/routes/api/v1 -v -run Test{Feature}
mage test:feature  # Ensure no regressions

# Verify explicit permissions registered
grep "CollectRoute" pkg/routes/api/v1/*.go
```

### Step 5: Clean Up routes.go

```bash
# Final cleanup
vim pkg/routes/routes.go

# Should only contain:
# - Framework setup
# - apiv1.Register* calls
# - apiv2.Register* calls
# - Middleware configuration
# Target: <250 lines total
```

### Step 6: Validate Architectural Consistency

```bash
# Count route patterns
echo "Modern routes:"
ls pkg/routes/api/v1/*.go | wc -l  # Should be ~24+ files

echo "Legacy routes:"
grep "WebHandler" pkg/routes/routes.go | wc -l  # Should be 0

# Verify permission registration
go test ./pkg/models -v -run TestAPITokenPermissions
```

## Phase 3: Comprehensive Validation (Target: 2-3 days)

### Step 1: Automated Test Parity Analysis

1. **Compare Test Suites**
   ```bash
   # Create comparison script
   find ../vikunja_original_main -name "*_test.go" > original_tests.txt
   find . -name "*_test.go" > refactored_tests.txt
   
   # Analyze differences
   diff original_tests.txt refactored_tests.txt
   ```

2. **Identify Missing Test Cases**
   ```bash
   # For each missing test, determine if:
   # - Test is obsolete (model logic moved to service)
   # - Test needs migration to service layer
   # - Test was accidentally lost
   ```

3. **Migrate Critical Tests**
   ```bash
   # Migrate important test cases to service layer
   # Adapt for new architecture patterns
   # Ensure equivalent coverage
   ```

### Step 2: Functional Parity Validation

1. **Execute Core Workflow Checklist**
   
   **Project Management Workflow:**
   ```bash
   # Original System Test:
   cd ../vikunja_original_main && ./vikunja web
   # 1. Create new project
   # 2. Add tasks to project  
   # 3. Assign users to tasks
   # 4. Mark tasks complete
   # 5. Verify project statistics
   
   # Refactored System Test:  
   cd ../vikunja && ./vikunja web
   # Execute identical workflow
   # Compare results step-by-step
   ```

   **Task Management Workflow:**
   ```bash
   # Test task creation, editing, deletion
   # Verify labels, attachments, assignees
   # Check related tasks functionality
   # Validate task filtering and search
   ```

   **Permission Workflow:**
   ```bash
   # Test project sharing
   # Verify user permissions
   # Check team permissions
   # Test link sharing functionality
   ```

2. **Document Any Differences**
   ```bash
   # If differences found:
   # - Original system behavior takes precedence
   # - Fix refactored system to match
   # - Re-test until identical
   ```

### Step 3: Architectural Review

1. **AI Analysis**
   ```bash
   # Systematic codebase analysis
   # Check architectural patterns
   # Verify constitutional compliance
   # Generate violation report
   ```

2. **Human Approval Process**
   ```bash
   # Review AI analysis results
   # Validate architectural decisions
   # Approve final implementation
   # Document sign-off
   ```

### Step 4: Final Quality Gates

1. **Test Coverage Validation**
   ```bash
   go test -coverprofile=coverage.out ./pkg/services/...
   go tool cover -html=coverage.out
   
   # Verify: ≥90% service layer coverage
   # Verify: ≥80% overall backend coverage
   ```

2. **Performance Validation**
   ```bash
   # Load test critical endpoints
   # Verify <200ms p95 response times
   # Confirm no performance regression
   ```

## Success Criteria

✅ **Phase 1 Complete**: All tests pass, UI bugs fixed, functional parity demonstrated  
✅ **Phase 2.1-2.3 Complete**: All 18 features refactored following architectural patterns  
✅ **Phase 2.4 Complete**: All routes migrated to declarative pattern, architectural consistency achieved  
✅ **Phase 3 Complete**: Test parity confirmed, functional validation passed, architectural approval granted

## Rollback Plan

If issues arise:
1. **Immediate**: Switch to `vikunja_original_main` reference
2. **Analysis**: Identify specific failing component  
3. **Targeted Fix**: Address specific issue without full rollback
4. **Validation**: Re-run affected test suites

## Emergency Contacts

- **Architecture Questions**: Refer to `REFACTORING_GUIDE.md`
- **Constitutional Compliance**: Check `constitution.md`  
- **Pattern Examples**: See existing refactored services (e.g., Project delete)

**Estimated Total Duration**: 10-14 days  
**Critical Path**: Phase 1 stabilization → Phase 2 dependency order → Phase 2.4 route modernization → Phase 3 validation