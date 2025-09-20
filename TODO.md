# TODO List for E2E Test Fixes (Latest: September 20, 2025)

## Current Status: Link Share & Table View Test Fixes

### 📋 Active Tasks

#### Link Share Tests (sharing/linkShare.spec.ts)
- [ ] **Check task list DOM structure** in shared project views
- [ ] **Verify CSS selectors** - `.tasks` element existence
- [ ] **Test share authentication flow** with hash tokens
- [ ] **Fix missing task rendering** in share context

#### Table View Tests (project/project-view-table.spec.ts)
- [ ] **Update API intercept patterns** for `/projects/1/views/3/tasks**`
- [ ] **Verify table view ID** - check if view ID `3` is correct
- [ ] **Fix loadTasks route matching** in cy.wait() calls
- [ ] **Test task navigation functionality**

### 🔍 Investigation Tasks
- [ ] **Find current CSS classes** used for task lists
- [ ] **Check project view IDs** and routing structure
- [ ] **Verify API endpoint patterns** in network requests
- [ ] **Test share token processing** in frontend

### 🧪 Testing & Validation
- [ ] **Run fixed tests locally** with both servers
- [ ] **Check lint/typecheck** passes after changes
- [ ] **Verify no regressions** in passing tests
- [ ] **Commit changes** with conventional commit messages
- [ ] **Push and monitor CI** results

## Recent Failures Analysis

### From GitHub Actions Run 17883608960:
1. **sharing/linkShare.spec.ts**: 2/3 tests failing - `.tasks` element not found
2. **project/project-view-table.spec.ts**: 2/3 tests failing - API intercept timeout
3. **Timeouts**: Jobs timing out after 20+ minutes due to hanging tests

### Root Causes Identified:
- **CSS Selector Changes**: `.tasks` class may have been renamed/restructured
- **API Pattern Changes**: URL patterns for table view requests may be outdated
- **Share Authentication**: Hash-based token processing may not be working correctly

## ✅ Previously Completed

### Major Fixes Applied:
- [x] **ViewKind Type Conversion** - Fixed numeric vs string format mismatch
- [x] **Subtask Relation Conflicts** - Resolved 409 API errors
- [x] **Router Parameter Safety** - Fixed unsafe parseInt() usage
- [x] **Button Rendering Issues** - Added missing hasPrimaryAction props
- [x] **Project Creation Tests** - Team and project creation flows working
- [x] **Task List Rendering** - Core DOM rendering issues resolved

### Static Analysis Status:
- [x] **ESLint**: Clean (0 errors)
- [x] **TypeScript**: Clean (0 type errors)
- [x] **Unit Tests**: 690/690 passing
- [x] **Build**: Frontend builds successfully

## 🎯 Current Focus

### Priority 1: Link Share Tests
The `.tasks` element is critical for multiple tests - need to identify:
1. What CSS class is actually used for task containers now
2. Whether tasks are rendering at all in shared projects
3. If share authentication is working properly

### Priority 2: Table View API Integration
The `loadTasks` route intercept is not matching actual requests - need to:
1. Check what API calls are actually made for table view
2. Update intercept patterns to match current endpoints
3. Verify view ID `3` is correct for table view

### Expected Impact:
Fixing these two test files should resolve ~6 failing tests and eliminate the timeout issues that are causing CI jobs to hang for 20+ minutes.

## Risk Management
- Making minimal, targeted changes only
- Testing each fix locally before committing
- Focusing on test fixes, not application logic changes
- Maintaining all existing functionality