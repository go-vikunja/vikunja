# TypeScript and Linting Issues Analysis

## Summary
Found 187 total issues (179 errors, 8 warnings) in the Vikunja frontend codebase:

## Issue Categories

### 1. TypeScript `any` Types (Main Focus)
- **Count**: ~150+ errors
- **Rule**: `@typescript-eslint/no-explicit-any`
- **Files affected**: Nearly all component and service files
- **Priority**: High - These compromise type safety

### 2. Vue Component Issues
- **Missing prop defaults**: Button.vue (4 warnings)
- **Reactivity loss**: ProjectList.vue, ViewEditForm.vue
- **Deprecated filters**: EditTeam.vue
- **Event hyphenation**: ProjectSettingsWebhooks.vue, EditTeam.vue

### 3. Code Quality Issues
- **Unused variables**: Multiple files using `IAbstract` and other imports
- **Indentation**: AddTask.vue, Attachments.vue, ShowTasks.vue
- **Missing trailing commas**: ProjectList.vue, QuickActions.vue
- **Lexical declarations**: QuickActions.vue

### 4. Vue Template Issues
- **HTML indentation**: ShowTasks.vue
- **Closing bracket newlines**: ShowTasks.vue

## Files with Most Issues (Priority Order)
1. **UserTeam.vue** - 13+ `any` type errors
2. **TaskService.ts** - Multiple service-level type issues
3. **QuickActions.vue** - Mixed issues including `any` types
4. **TaskDetailView.vue** - 5+ `any` type errors
5. **Various user settings views** - Multiple `any` type errors

## Fix Strategy
1. **Phase 1**: Fix Button.vue prop defaults (quick wins)
2. **Phase 2**: Systematically replace `any` types with proper TypeScript interfaces
3. **Phase 3**: Fix Vue reactivity and template issues
4. **Phase 4**: Clean up unused imports and variables
5. **Phase 5**: Fix formatting and indentation issues

## Notes
- TypeScript compilation is already passing (pnpm typecheck ✅)
- Issues are primarily ESLint/style related, not fundamental type errors
- Need to run tests after each batch to ensure functionality is preserved