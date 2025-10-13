import type {ITask} from '@/modelTypes/ITask'

/**
 * Determines if a task should be displayed in the List view.
 *
 * Subtasks are hidden only when their parent task is also in the current view
 * (same project). Cross-project subtasks remain visible.
 *
 * @param task - The task to check
 * @param allTasksInView - All tasks currently visible in the view
 * @returns true if the task should be shown, false if it should be hidden
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
