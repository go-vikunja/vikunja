# TypeScript Errors Analysis - Vikunja Frontend

## Categories of Errors Found

### 1. Kanban Store Issues (`src/stores/kanban.ts`)
- Type mismatch between `ITask` and `IBucket` arrays
- Undefined index type errors
- Parameter type issues (implicit `any`)
- Argument type mismatches

### 2. Task List Composable (`src/composables/useTaskList.ts`)
- Type `(ITask | IBucket)[]` not assignable to `ITask[]`

### 3. Component Type Issues
- **TaskDetailView.vue**: Many issues with readonly vs mutable types, missing properties, implicit `any` parameters
- **EditTeam.vue**: `IUser | undefined` type issues, missing properties
- **Avatar.vue**: `Blob | null` not assignable to `IAvatar`
- **DataExportDownload.vue**: Property access on `never` type

### 4. General Patterns
- Readonly/immutable types being assigned to mutable types
- Null/undefined handling issues
- Missing required properties when creating objects
- Implicit `any` parameter types
- Type union issues where specific types are expected

## Priority Areas to Fix

1. **High Priority - Core Store Issues**: Kanban store has fundamental type issues
2. **Medium Priority - Component Issues**: Various component type safety issues
3. **Low Priority - Parameter Types**: Implicit `any` parameter types

## Strategy
- Fix core type definitions first
- Handle null/undefined cases properly
- Add explicit typing where implicit `any` occurs
- Ensure proper type guards for union types