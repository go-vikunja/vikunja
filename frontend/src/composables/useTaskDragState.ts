import {shallowRef} from 'vue'
import {createSharedComposable} from '@vueuse/core'
import type {TaskResponse} from '@/client/queries/tasks'

export const useTaskDragState = createSharedComposable(() => {
	const draggedTask = shallowRef<TaskResponse | null>(null)
	return {draggedTask, setDraggedTask: (task: TaskResponse | null) => { draggedTask.value = task }}
})
