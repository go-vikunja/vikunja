# TypeScript Issues Analysis - Vikunja Frontend

## Summary
Found 182 TypeScript errors across various components and views. The issues fall into several categories:

## Issue Categories

### 1. Type Assignment Issues
- `ITask` not assignable to `Record<string, unknown>`
- Missing properties in type assignments
- `undefined` not assignable to specific types
- `null` not assignable to specific types

### 2. Missing Properties
- Missing required properties in interfaces (e.g., `maxPermission` in `ITaskReminder`)
- Properties that don't exist on types (e.g., `oidcId` on `ITeam`)

### 3. Null/Undefined Safety Issues
- Properties possibly `null` or `undefined` without proper checks
- Missing null checks before property access

### 4. Implicit Any Types
- Parameters with implicit `any` type
- Elements with implicit `any` type due to dynamic indexing

### 5. Function Call Issues
- Wrong number of arguments passed to functions
- Type mismatches in function parameters

## Key Files with Issues

### Components
- `src/components/tasks/partials/RelatedTasks.vue` - 2 errors
- `src/components/tasks/partials/Reminders.story.vue` - 4 errors
- `src/components/tasks/partials/Reminders.vue` - 2 errors
- `src/components/tasks/partials/RepeatAfter.vue` - 1 error
- `src/components/tasks/partials/SingleTaskInProject.vue` - 15+ errors
- `src/components/input/editor.vue` - Multiple errors
- `src/components/misc/subscription.vue` - Multiple errors

### Views
- `src/views/tasks/TaskDetailView.vue` - 10+ errors
- `src/views/teams/EditTeam.vue` - 10+ errors
- `src/views/teams/NewTeam.vue` - 1 error
- `src/views/user/DataExportDownload.vue` - 1 error
- `src/views/user/settings/Avatar.vue` - 1 error

### Services & Stores
- `src/services/caldav.ts` - Multiple errors
- `src/stores/tasks.ts` - Multiple errors
- `src/stores/projects.ts` - Multiple errors

## Next Steps
1. Start with interface definitions to ensure proper typing
2. Fix null/undefined safety issues with proper checks
3. Address missing properties and type mismatches
4. Update function signatures and calls
5. Add explicit types where implicit any is detected

## Priority Order
1. Interface and type definition fixes
2. Null safety improvements
3. Missing property additions
4. Function signature corrections
5. Implicit any type fixes