import {computed, toValue, type MaybeRefOrGetter} from 'vue'
import {useStorage} from '@vueuse/core'

import type {ITask} from '@/modelTypes/ITask'

// Persist the collapsed state across reloads. Only collapsed tasks are kept so the entry
// does not grow with every task that was ever expanded.
type CollapsedTasks = {[key: ITask['id']]: true}

const collapsedTasks = useStorage<CollapsedTasks>('collapsedSubtasks', {})

function setCollapsed(taskId: ITask['id'], collapsed: boolean) {
	if (collapsed) {
		collapsedTasks.value[taskId] = true
		return
	}

	delete collapsedTasks.value[taskId]
}

export function useSubtasksCollapsed(taskId: MaybeRefOrGetter<ITask['id']>) {
	return computed({
		get: () => collapsedTasks.value[toValue(taskId)] === true,
		set: (collapsed: boolean) => setCollapsed(toValue(taskId), collapsed),
	})
}

export function useCollapseAllSubtasks(tasks: MaybeRefOrGetter<ITask[]>) {
	const tasksWithSubtasks = computed(() => toValue(tasks).filter(t => t.relatedTasks?.subtask?.length))

	const hasTasksWithSubtasks = computed(() => tasksWithSubtasks.value.length > 0)
	const allCollapsed = computed(() => hasTasksWithSubtasks.value &&
		tasksWithSubtasks.value.every(t => collapsedTasks.value[t.id] === true))

	function toggleAll() {
		const collapse = !allCollapsed.value
		tasksWithSubtasks.value.forEach(t => setCollapsed(t.id, collapse))
	}

	return {
		allCollapsed,
		hasTasksWithSubtasks,
		toggleAll,
	}
}
