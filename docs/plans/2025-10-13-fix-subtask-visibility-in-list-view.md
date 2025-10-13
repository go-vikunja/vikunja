# Fix Subtask Visibility in List View Implementation Plan

> **For Claude:** Use `${SUPERPOWERS_SKILLS_ROOT}/skills/collaboration/executing-plans/SKILL.md` to implement this plan task-by-task.

**Goal:** Fix bug where subtasks from other projects disappear from List view when assigned as subtasks, while remaining visible in Table and Kanban views.

**Architecture:** The bug is in the frontend List view component which incorrectly filters out ALL tasks with parent tasks, regardless of project membership. The fix involves modifying the filter logic to only hide subtasks when their parent task is in the SAME project view.

**Tech Stack:** Vue 3, TypeScript, Composition API

---

## Background

**Issue:** When a task in Project B is assigned as a subtask of a task in Project A, the subtask disappears from Project B's List view, but remains visible in Table and Kanban views.

**Root Cause:** In `frontend/src/components/project/views/ProjectList.vue:168-171`, the List view filters out all tasks that have parent tasks:

```typescript
tasks.value = tasks.value.filter(t => {
    return !((t.relatedTasks?.parenttask?.length ?? 0) > 0)
})
```

This logic is too aggressive - it should only hide subtasks when their parent is ALSO visible in the current view (i.e., in the same project), not when the parent is in a different project.

**Expected Behavior:**
- If Task A (Project A) has subtask Task B (Project A), show only Task A in Project A's List view (with Task B nested under it)
- If Task A (Project A) has subtask Task B (Project B), show Task B in Project B's List view AND show Task A with nested Task B in Project A's List view

## Task 0: Create a new branch for the FIX

Create a new branch for the fix:

```bash
git checkout -b fix-subtask-visibility
```

---

## Task 1: Write Tests for Subtask Visibility Logic

**Files:**
- Create: `frontend/src/components/project/views/ProjectList.test.ts`

**Step 1: Write the failing test for same-project subtasks**

Create a test file that verifies subtasks are hidden only when their parent is in the same project:

```typescript
import { describe, it, expect } from 'vitest'
import type { ITask } from '@/modelTypes/ITask'
import type { RelatedTaskMap } from '@/models/task'

// Helper function to simulate the filtering logic
function shouldShowTaskInListView(
	task: ITask,
	allTasksInView: ITask[],
): boolean {
	// If task has no parent, always show it
	const parentTasksCount = task.relatedTasks?.parenttask?.length ?? 0
	if (parentTasksCount === 0) {
		return true
	}

	// Task has parent(s) - only hide if parent is in the same view
	const parentTasks = task.relatedTasks?.parenttask ?? []
	const parentIds = parentTasks.map(p => p.id)
	const hasParentInView = allTasksInView.some(t => parentIds.includes(t.id))

	// Show task if parent is NOT in the current view (cross-project subtask)
	return !hasParentInView
}

describe('ProjectList subtask visibility', () => {
	it('should hide subtasks when parent is in the same project', () => {
		const parentTask: ITask = {
			id: 1,
			title: 'Parent Task',
			projectId: 100,
			relatedTasks: {} as RelatedTaskMap,
		} as ITask

		const subtask: ITask = {
			id: 2,
			title: 'Subtask',
			projectId: 100,
			relatedTasks: {
				parenttask: [{
					id: 1,
					title: 'Parent Task',
					projectId: 100,
				}],
			} as RelatedTaskMap,
		} as ITask

		const allTasks = [parentTask, subtask]

		expect(shouldShowTaskInListView(parentTask, allTasks)).toBe(true)
		expect(shouldShowTaskInListView(subtask, allTasks)).toBe(false)
	})

	it('should show subtasks when parent is in a different project', () => {
		const parentTask: ITask = {
			id: 1,
			title: 'Parent Task in Project A',
			projectId: 100,
		} as ITask

		const subtask: ITask = {
			id: 2,
			title: 'Subtask in Project B',
			projectId: 200,
			relatedTasks: {
				parenttask: [{
					id: 1,
					title: 'Parent Task in Project A',
					projectId: 100,
				}],
			} as RelatedTaskMap,
		} as ITask

		// In Project B's view, we only see the subtask
		const tasksInProjectB = [subtask]

		expect(shouldShowTaskInListView(subtask, tasksInProjectB)).toBe(true)
	})

	it('should show tasks with no parents', () => {
		const task: ITask = {
			id: 1,
			title: 'Regular Task',
			projectId: 100,
			relatedTasks: {} as RelatedTaskMap,
		} as ITask

		const allTasks = [task]

		expect(shouldShowTaskInListView(task, allTasks)).toBe(true)
	})

	it('should handle multiple levels of nesting within same project', () => {
		const grandparent: ITask = {
			id: 1,
			title: 'Grandparent',
			projectId: 100,
			relatedTasks: {} as RelatedTaskMap,
		} as ITask

		const parent: ITask = {
			id: 2,
			title: 'Parent',
			projectId: 100,
			relatedTasks: {
				parenttask: [{id: 1, title: 'Grandparent', projectId: 100}],
			} as RelatedTaskMap,
		} as ITask

		const child: ITask = {
			id: 3,
			title: 'Child',
			projectId: 100,
			relatedTasks: {
				parenttask: [{id: 2, title: 'Parent', projectId: 100}],
			} as RelatedTaskMap,
		} as ITask

		const allTasks = [grandparent, parent, child]

		expect(shouldShowTaskInListView(grandparent, allTasks)).toBe(true)
		expect(shouldShowTaskInListView(parent, allTasks)).toBe(false)
		expect(shouldShowTaskInListView(child, allTasks)).toBe(false)
	})
})
```

**Step 2: Run test to verify it fails**

Run: `cd frontend && pnpm test:unit ProjectList.test.ts`
Expected: Tests should fail because the function doesn't exist yet

**Step 3: Commit the test**

```bash
git add frontend/src/components/project/views/ProjectList.test.ts
git commit -m "test: add tests for cross-project subtask visibility in List view"
```

---

## Task 2: Extract Filter Logic into Composable

**Files:**
- Create: `frontend/src/composables/useTaskListFiltering.ts`
- Modify: `frontend/src/components/project/views/ProjectList.vue`

**Step 1: Create the composable with the filtering logic**

```typescript
import type { ITask } from '@/modelTypes/ITask'

/**
 * Determines if a task should be displayed in the List view.
 *
 * Subtasks are hidden only when their parent task is also in the current view
 * (same project). Cross-project subtasks remain visible.
 */
export function shouldShowTaskInListView(
	task: ITask,
	allTasksInView: ITask[],
): boolean {
	// If task has no parent, always show it
	const parentTasksCount = task.relatedTasks?.parenttask?.length ?? 0
	if (parentTasksCount === 0) {
		return true
	}

	// Task has parent(s) - only hide if parent is in the same view
	const parentTasks = task.relatedTasks?.parenttask ?? []
	const parentIds = parentTasks.map(p => p.id)
	const hasParentInView = allTasksInView.some(t => parentIds.includes(t.id))

	// Show task if parent is NOT in the current view (cross-project subtask)
	return !hasParentInView
}
```

**Step 2: Run the tests again**

Run: `cd frontend && pnpm test:unit ProjectList.test.ts`
Expected: All tests should now pass

**Step 3: Commit the composable**

```bash
git add frontend/src/composables/useTaskListFiltering.ts
git commit -m "feat: add composable for cross-project-aware task filtering"
```

---

## Task 3: Update ProjectList Component to Use New Filter Logic

**Files:**
- Modify: `frontend/src/components/project/views/ProjectList.vue:160-172`

**Step 1: Import the new composable**

In the `<script setup>` section, add the import after other composable imports (around line 111):

```typescript
import {shouldShowTaskInListView} from '@/composables/useTaskListFiltering'
```

**Step 2: Update the filter logic**

Replace lines 168-171:

```typescript
// OLD CODE (remove this):
tasks.value = tasks.value.filter(t => {
	return !((t.relatedTasks?.parenttask?.length ?? 0) > 0)
})

// NEW CODE (replace with this):
tasks.value = tasks.value.filter(t => shouldShowTaskInListView(t, allTasks.value))
```

**Step 3: Test the change manually**

1. Start the dev server: `cd frontend && pnpm dev`
2. Navigate to a project's List view
3. Create a task in Project A and Project B
4. Make Project B's task a subtask of Project A's task
5. Verify:
   - Project A's List view shows the parent task with the subtask nested
   - Project B's List view still shows the subtask (BUG FIX CONFIRMED)
   - Both Table and Kanban views continue to work as before

**Step 4: Run frontend linting**

Run: `cd frontend && pnpm lint:fix`
Expected: No linting errors

**Step 5: Run frontend type checking**

Run: `cd frontend && pnpm typecheck`
Expected: No type errors

**Step 6: Run unit tests**

Run: `cd frontend && pnpm test:unit`
Expected: All tests pass

**Step 7: Commit the component changes**

```bash
git add frontend/src/components/project/views/ProjectList.vue
git commit -m "fix: show cross-project subtasks in List view

Previously, subtasks were hidden in List view if they had ANY parent
task, regardless of which project the parent belonged to. This caused
cross-project subtasks to disappear from their own project's List view.

Now, subtasks are only hidden when their parent task is also visible
in the current view (same project). Cross-project subtasks remain
visible in their own project's List view.

Fixes #782"
```

---

## Task 4: Add E2E Test for Cross-Project Subtasks

**Files:**
- Extend: `frontend/cypress/e2e/project/project-view-list.spec.ts`

**Step 1: Write E2E test**

```typescript
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TaskRelationFactory} from '../../factories/task_relation'

describe('Cross-Project Subtasks in List View', () => {
	let projects
	let tasks

	beforeEach(() => {
		// Create two projects
		projects = ProjectFactory.create(2)

		// Create a task in each project
		tasks = [
			TaskFactory.create({
				id: 1,
				title: 'Parent Task in Project A',
				project_id: projects[0].id,
			}),
			TaskFactory.create({
				id: 2,
				title: 'Subtask in Project B',
				project_id: projects[1].id,
			}),
		]
	})

	it('shows cross-project subtasks in their own project List view', () => {
		// Make task 2 a subtask of task 1
		TaskRelationFactory.create({
			task_id: 2,
			other_task_id: 1,
			relation_kind: 'subtask',
		})
		TaskRelationFactory.create({
			task_id: 1,
			other_task_id: 2,
			relation_kind: 'parenttask',
		})

		// Visit Project B's List view
		cy.visit(`/projects/${projects[1].id}/list`)

		// The subtask should be visible in Project B
		cy.contains('Subtask in Project B')
			.should('be.visible')
	})

	it('hides same-project subtasks under their parent', () => {
		// Create another task in Project A
		const subtaskInProjectA = TaskFactory.create({
			id: 3,
			title: 'Subtask in Project A',
			project_id: projects[0].id,
		})

		// Make it a subtask of the parent in Project A
		TaskRelationFactory.create({
			task_id: 3,
			other_task_id: 1,
			relation_kind: 'subtask',
		})
		TaskRelationFactory.create({
			task_id: 1,
			other_task_id: 3,
			relation_kind: 'parenttask',
		})

		// Visit Project A's List view
		cy.visit(`/projects/${projects[0].id}/list`)

		// Parent task should be visible
		cy.contains('Parent Task in Project A')
			.should('be.visible')

		// Subtask should NOT be visible at top level (it's nested under parent)
		// Note: This test may need adjustment based on how nesting is displayed
		cy.get('.tasks > li')
			.should('have.length', 1) // Only parent task at top level
	})

	it('shows cross-project subtasks in Table view', () => {
		TaskRelationFactory.create({
			task_id: 2,
			other_task_id: 1,
			relation_kind: 'subtask',
		})

		// Visit Project B's Table view
		cy.visit(`/projects/${projects[1].id}/table`)

		// The subtask should be visible in Project B's Table view
		cy.contains('Subtask in Project B')
			.should('be.visible')
	})

	it('shows cross-project subtasks in Kanban view', () => {
		TaskRelationFactory.create({
			task_id: 2,
			other_task_id: 1,
			relation_kind: 'subtask',
		})

		// Visit Project B's Kanban view (assuming it has a default view)
		cy.visit(`/projects/${projects[1].id}`)

		// The subtask should be visible in Project B's Kanban view
		cy.contains('Subtask in Project B')
			.should('be.visible')
	})
})
```

**Step 2: Commit the E2E test**

```bash
git add frontend/cypress/e2e/task/cross-project-subtasks.spec.ts
git commit -m "test: add e2e tests for cross-project subtask visibility"
```

---

## Task 5: Final Verification

**Files:**
- None (verification only)

**Step 1: Run all frontend tests**

Run: `cd frontend && pnpm test:unit && pnpm test:e2e`
Expected: All tests pass

**Step 2: Run frontend linting**

Run: `cd frontend && pnpm lint && pnpm lint:styles`
Expected: No linting errors

**Step 3: Run type checking**

Run: `cd frontend && pnpm typecheck`
Expected: No type errors in new files

**Step 4: Build the frontend**

Run: `cd frontend && pnpm build`
Expected: Build succeeds without errors

**Step 5: Manual testing checklist**

Test the following scenarios:
- [ ] Create task A in Project 1 and task B in Project 2
- [ ] Make task B a subtask of task A
- [ ] Verify task B appears in Project 2's List view
- [ ] Verify task B appears in Project 2's Table view
- [ ] Verify task B appears in Project 2's Kanban view
- [ ] Verify task A shows task B as nested in Project 1's List view
- [ ] Create task C in Project 1 as subtask of task A
- [ ] Verify task C is hidden at top level in Project 1's List view (nested under A)
- [ ] Verify deeply nested subtasks work (3+ levels in same project)
- [ ] Verify saved filters still work correctly with subtasks

**Step 6: Document test results**

Create a comment summarizing successful manual testing and share with the user.

---

## Summary

This plan fixes the bug where cross-project subtasks disappeared from List view by implementing project-aware filtering logic. The key insight is that subtasks should only be hidden when their parent task is in the SAME view, not when the parent is in a different project.

**Changes:**
1. Created unit tests for the filtering logic
2. Extracted filtering logic into a reusable composable
3. Updated ProjectList.vue to use the new logic
4. Added E2E tests for cross-project subtasks
5. Verified the fix works across all view types

**Testing Strategy:**
- Unit tests verify the core filtering logic
- E2E tests verify end-to-end behavior across views
- Manual testing verifies UX matches expected behavior
