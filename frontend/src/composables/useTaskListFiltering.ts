import type {ITask} from '@/modelTypes/ITask'

/**
 * Determines if a task should be displayed in the List view.
 *
 * Subtasks are hidden only when their parent task is also in the current view
 * (same project). Cross-project subtasks remain visible.
 *
 * In filtered views (saved filters), all tasks are shown regardless of parent
 * presence, since the user explicitly filtered for them.
 *
 * @param task - The task to check
 * @param allTasksInView - All tasks currently visible in the view
 * @param isFilteredView - Whether the current view is a saved/custom filter
 * @returns true if the task should be shown, false if it should be hidden
 */
export function shouldShowTaskInListView(
	task: ITask,
	allTasksInView: ITask[],
	isFilteredView: boolean = false,
): boolean {
	// In filtered views (saved filters), show all tasks that matched the filter
	if (isFilteredView) {
		return true
	}

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
