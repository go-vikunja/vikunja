# Remaining TypeScript Issues Plan

Based on the current typecheck output and previous analysis, here's the plan to finish resolving all TypeScript issues:

## Current Status
- Previous work has resolved ~60-70% of TypeScript errors
- Estimated ~150-200 errors remain (from typecheck output)
- All tests are currently passing
- Major architectural issues have been resolved

## Remaining Issue Categories (Priority Order)

### 1. HIGH PRIORITY - ViewEditForm.vue
**Errors:** 6 errors related to filter properties
- Line 68: `IFilters | undefined` not assignable to `IFilters`
- Line 196: `view.filter` possibly undefined
- Line 270, 278: Objects possibly undefined
- Line 293: Missing properties in filter object

**Fix Strategy:** Add proper null checks and initialize IFilters with required properties

### 2. HIGH PRIORITY - QuickActions.vue
**Errors:** ~15 errors related to type conversions and null safety
- String to number conversion issues
- Null/undefined access on action types
- Property access on wrong types ('id', 'title' on ACTION_TYPE)
- Complex computed property type mismatches

**Fix Strategy:** Add proper type guards, fix conversions, handle null cases

### 3. MEDIUM PRIORITY - Union Type Complexity
**Files:** Button.vue, CreateEdit.vue, Dropdown.vue
**Errors:** TS2590 - Union type too complex to represent
**Note:** These are Vue compiler edge cases, functionality works

**Fix Strategy:** Simplify prop type definitions or use interface separation

### 4. MEDIUM PRIORITY - Authentication Forms
**Files:** Login.vue, Register.vue, PasswordReset.vue, RequestPasswordReset.vue
**Errors:** Missing properties, unknown error handling, validation issues
- Missing 'totpPasscode' property
- Error type handling (unknown 'e' types)
- Missing modelValue properties

**Fix Strategy:** Add missing interface properties, proper error typing

### 5. MEDIUM PRIORITY - Team Management Components
**Files:** EditTeam.vue, ListTeams.vue, NewTeam.vue
**Errors:** Type mismatches in team interfaces
- Missing ITeamMember properties
- Property access on 'never' types
- Incorrect type assignments

**Fix Strategy:** Fix interface definitions and property mappings

### 6. LOW PRIORITY - Scattered Issues
**Various Files:** Null safety, missing properties, type casting
**Examples:** DataExportDownload.vue, Avatar.vue, etc.

## Execution Plan

### Batch 1: ViewEditForm.vue (Immediate Impact)
- Fix filter-related null checks and initialization
- Test and commit

### Batch 2: QuickActions.vue (Complex but High Impact)
- Fix type conversions and null safety
- Add proper type guards
- Test and commit

### Batch 3: Union Type Issues (Lower Risk)
- Simplify complex union types in components
- Test and commit

### Batch 4: Authentication & Teams (User-Facing)
- Fix auth form issues
- Fix team management types
- Test and commit

### Batch 5: Cleanup (Final Pass)
- Address remaining scattered issues
- Final typecheck verification
- Test and commit

## Success Criteria
- `pnpm typecheck` returns 0 errors
- All unit tests pass (`pnpm test:unit`)
- All e2e tests pass (`pnpm test:e2e`)
- No functional regressions introduced