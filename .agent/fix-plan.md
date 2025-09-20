# TypeScript Fix Plan

## Batch 1: Core Data Structure Issues (Readonly/Mutable)
Focus on the most critical issues first - readonly object assignments that break core functionality.

### Files to fix:
1. `src/views/project/ListProjects.vue:33` - readonly array assignment
2. `src/views/project/NewProject.vue:41,84` - readonly object assignments
3. `src/views/project/ProjectView.vue` - multiple readonly/type issues

### Strategy:
- Use type assertions where safe
- Add proper type guards
- Ensure API response types match expected interfaces

## Batch 2: Project and Task Type Issues
Continue with project and task-related type problems.

### Files to fix:
1. Remaining ProjectView.vue issues
2. ShowTasks.vue - date/array handling issues
3. Task-related components with type mismatches

## Batch 3: Component and Props Issues
Fix component prop and event handler type issues.

### Files to fix:
1. Various component prop missing/incorrect types
2. Event handler type mismatches
3. Generic type parameter issues

## Batch 4: Settings and Minor Issues
Handle remaining settings pages and minor type issues.

### Files to fix:
1. Project settings components
2. User settings components
3. Webhook and other settings

## Testing Strategy:
- Run `pnpm test:unit` after each batch
- Run `pnpm test:e2e` after major batches
- Commit after each successful batch of fixes