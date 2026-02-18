import {useI18n} from 'vue-i18n'

import type {ITask} from '@/modelTypes/ITask'
import {useTaskStore} from '@/stores/tasks'
import {useProjectStore} from '@/stores/projects'
import {success, error} from '@/message'

/**
 * Finds a project ID from elements at a given mouse position.
 * Searches through elements under the mouse and their parents for data-project-id attribute.
 */
function findProjectIdAtPosition(mouseX: number, mouseY: number): number | null {
	// Validate coordinates are finite numbers (required by elementsFromPoint)
	// This prevents errors on Firefox mobile where touch events may provide invalid coordinates
	if (!Number.isFinite(mouseX) || !Number.isFinite(mouseY)) {
		return null
	}

	const elementsUnderMouse = document.elementsFromPoint(mouseX, mouseY)

	for (const el of elementsUnderMouse) {
		if (!(el instanceof HTMLElement)) {
			continue
		}

		const withProjectId =
			el.dataset?.projectId != null
				? el
				: el.closest('[data-project-id]') as HTMLElement | null

		const projectId = withProjectId?.dataset.projectId
		if (projectId) {
			const parsed = parseInt(projectId, 10)
			if (!Number.isNaN(parsed)) {
				return parsed
			}
		}
	}

	return null
}

export interface TaskDragToProjectResult {
	moved: boolean
	targetProjectId: number | null
}

/**
 * Composable for handling task drag-and-drop to sidebar projects.
 *
 * Provides functionality to:
 * - Detect when a task is dropped over a sidebar project
 * - Move the task to the target project
 * - Show success/error notifications
 *
 * @returns Functions for handling drag start and checking for project drops
 */
export function useTaskDragToProject() {
	const {t} = useI18n({useScope: 'global'})
	const taskStore = useTaskStore()
	const projectStore = useProjectStore()

	/**
	 * Attempts to move a dragged task to a project at the given mouse position.
	 * Should be called in the drag end handler.
	 *
	 * @param e - The drag event with originalEvent containing mouse coordinates
	 * @param onSuccess - Optional callback called after successful move (e.g., to update local state)
	 * @returns Result indicating if task was moved and to which project
	 */
	async function handleTaskDropToProject(
		e: { originalEvent?: MouseEvent },
		onSuccess?: (task: ITask, targetProjectId: number) => void,
	): Promise<TaskDragToProjectResult> {
		const draggedTask = taskStore.draggedTask

		if (!draggedTask || !e.originalEvent) {
			taskStore.setDraggedTask(null)
			return {moved: false, targetProjectId: null}
		}

		const mouseX = e.originalEvent.clientX
		const mouseY = e.originalEvent.clientY
		const targetProjectId = findProjectIdAtPosition(mouseX, mouseY)

		if (!targetProjectId || targetProjectId <= 0 || targetProjectId === draggedTask.projectId) {
			taskStore.setDraggedTask(null)
			return {moved: false, targetProjectId}
		}

		const targetProject = projectStore.projects[targetProjectId]

		try {
			await taskStore.update({
				...draggedTask,
				projectId: targetProjectId,
			})

			if (onSuccess) {
				onSuccess(draggedTask, targetProjectId)
			}

			success({message: t('task.movedToProject', {project: targetProject?.title || t('project.title')})})

			return {moved: true, targetProjectId}
		} catch (e) {
			error(e)
			return {moved: false, targetProjectId}
		} finally {
			// Always clears drag state - callers should not clear again
			taskStore.setDraggedTask(null)
		}
	}

	return {
		handleTaskDropToProject,
	}
}
