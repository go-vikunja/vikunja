# T-PERM-014A Phase 3: High-Impact Helper Removals - Implementation Plan

**Date**: 2025-01-13  
**Status**: TODO  
**Priority**: HIGH  
**Estimated Time**: 1-1.5 days  
**Dependencies**: T-PERM-014A Phase 2  
**Parent Task**: T-PERM-014A  

## Scope Summary

Remove 4 remaining helper functions from models with **35+ total call sites**:
- `GetProjectSimpleByID()` - 25+ call sites (HIGH IMPACT)
- `GetProjectsMapByIDs()` - 3+ call sites
- `GetProjectsByIDs()` - 3+ call sites
- `GetTaskByIDSimple()` - 10+ call sites (HIGH IMPACT)

**Total Estimated Call Sites**: 35-40+

## Risk Assessment

**Circular Dependency Risk**: **HIGH**
- Many services may need ProjectService
- ProjectService may need other services
- Careful dependency analysis required

**Breaking Change Risk**: LOW  
- All functions already delegate to services
- Internal refactoring only

**Testing Effort**: **HIGH**
- 35+ call sites to update
- Multiple service files affected
- Integration testing critical

## Phase 3 Breakdown

### Category 2: Project Helpers (28+ call sites)

#### Function 1: GetProjectSimpleByID() - 25+ call sites

**Current Implementation**:
```go
// pkg/models/project.go
func GetProjectSimpleByID(s *xorm.Session, projectID int64) (project *Project, err error) {
	if GetProjectByIDSimpleFunc != nil {
		return GetProjectByIDSimpleFunc(s, projectID)
	}
	// Fallback...
	service := getProjectService()
	return service.GetByIDSimple(s, projectID)
}
```

**Usage Analysis Required**:
```bash
# Find all usages
grep -r "GetProjectSimpleByID" vikunja/pkg --include="*.go" | wc -l
# Expected: 25+ matches across services and models
```

**Strategy**:
1. Audit all call sites - categorize by file
2. For service layer: Add ProjectService dependencies where needed
3. For model layer: Use function pointer (already exists: `GetProjectByIDSimpleFunc`)
4. Check for circular dependencies
5. Update calls incrementally, test after each major change

**High Risk Areas**:
- Services that don't currently have ProjectService dependency
- Model methods that call this helper
- Test code that may need updates

#### Function 2: GetProjectsMapByIDs() - 3+ call sites

**Current Implementation**:
```go
// pkg/models/project.go
func GetProjectsMapByIDs(s *xorm.Session, projectIDs []int64) (projectMap map[int64]*Project, err error) {
	if GetProjectMapByIDsFunc != nil {
		return GetProjectMapByIDsFunc(s, projectIDs)
	}
	// Fallback...
	service := getProjectService()
	return service.GetMapByIDs(s, projectIDs)
}
```

**Strategy**:
- Similar to GetProjectSimpleByID
- Lower impact due to fewer call sites
- May have same circular dependency concerns

#### Function 3: GetProjectsByIDs() - 3+ call sites

**Current Implementation**:
```go
// pkg/models/project.go
func GetProjectsByIDs(s *xorm.Session, projectIDs []int64, showArchived bool) (projects []*Project, err error) {
	if GetProjectByIDsFunc != nil {
		return GetProjectByIDsFunc(s, projectIDs, showArchived)
	}
	// Fallback...
	service := getProjectService()
	return service.GetByIDs(s, projectIDs, showArchived)
}
```

**Strategy**:
- Similar approach as above
- Check if `showArchived` parameter complicates things

### Category 5: Task Helpers (10+ call sites)

#### Function 4: GetTaskByIDSimple() - 10+ call sites

**Current Implementation**:
```go
// pkg/models/tasks.go
func GetTaskByIDSimple(s *xorm.Session, taskID int64) (task *Task, err error) {
	if GetTaskByIDSimpleFunc != nil {
		return GetTaskByIDSimpleFunc(s, taskID)
	}
	// Fallback...
	panic("TaskService not initialized")
}
```

**Known Usage**:
- `pkg/services/attachment.go` - Permission checks
- `pkg/services/comment.go` - Permission checks
- `pkg/services/task.go` - Internal methods
- Model files - Various places

**Strategy**:
1. Add TaskService dependencies to services that need it
2. Update service layer calls first
3. Update model layer to use function pointer
4. Wire function pointer if not already done

**Circular Dependency Risk**: MEDIUM
- AttachmentService → TaskService (OK)
- CommentService → TaskService (already has it? check)
- TaskService → TaskService (no issue, internal)

## Implementation Order

### Step 1: Audit All Call Sites (0.25 days)
1. Run grep searches for each function
2. Categorize by file and layer (service vs model)
3. Document current dependencies
4. Identify circular dependency risks
5. Create detailed call site matrix

### Step 2: Project Helpers - Service Layer (0.5 days)
1. Identify services needing ProjectService dependency
2. Add dependencies (check for cycles!)
3. Update service layer calls (25+ sites)
4. Build and test incrementally

### Step 3: Project Helpers - Model Layer (0.25 days)
1. Verify function pointers exist and are wired
2. Update model layer calls
3. Test model compilation

### Step 4: Task Helpers (0.5 days)
1. Add TaskService dependencies where needed
2. Update service layer calls (10+ sites)
3. Update model layer calls
4. Wire function pointer if needed

### Step 5: Remove Helper Functions (0.1 days)
1. Remove all 4 helper functions from models
2. Verify no remaining references

### Step 6: Verification (0.25 days)
1. Build verification
2. Service tests
3. Baseline tests
4. Full test suite

## Verification Plan

```bash
cd /home/aron/projects/vikunja

# After each major change, verify:
go build ./pkg/models ./pkg/services

# After Step 2-4, run affected service tests:
export VIKUNJA_SERVICE_ROOTPATH=$(pwd)
go test ./pkg/services -run "TestProjectService|TestTaskService|TestAttachmentService|TestCommentService" -v

# After Step 5, verify no references remain:
grep -r "GetProjectSimpleByID\|GetProjectsMapByIDs\|GetProjectsByIDs\|GetTaskByIDSimple" pkg/

# Final verification:
go test ./pkg/services -run "TestPermissionBaseline" -v
mage test:all
```

## Success Criteria

- ✅ All 4 remaining helper functions removed
- ✅ All 35+ call sites updated
- ✅ Service dependencies added correctly
- ✅ No circular dependencies introduced
- ✅ Production code compiles cleanly
- ✅ All service tests pass
- ✅ Baseline tests pass (4/6 minimum, 5/6 after T-PERM-014A-FIX)
- ✅ Full test suite passes

## Known Challenges

1. **Scale**: 35+ call sites is significant
2. **Circular Dependencies**: High risk with ProjectService
3. **Integration**: Changes span multiple services
4. **Testing**: Many integration points to verify

## Mitigation Strategies

1. **Incremental Updates**: Update and test file-by-file
2. **Dependency Mapping**: Create visual dependency graph
3. **Rollback Points**: Commit after each successful file update
4. **Pair Review**: Have changes reviewed before moving to next file

## Follow-up Tasks Created

None yet - will create if circular dependency issues arise

## Notes

- ProjectService is heavily used - expect many dependencies to add
- TaskService already has many dependencies - less risky
- Consider creating helper document: DEPENDENCY-MAP.md showing all service dependencies
- May need to refactor service initialization order if cycles detected
