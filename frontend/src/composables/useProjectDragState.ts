import {ref} from 'vue'
import {createSharedComposable} from '@vueuse/core'

// Shared because a project can be dragged from one sidebar list into another.
export const useProjectDragState = createSharedComposable(() => {
	const isDraggingProject = ref(false)
	return {isDraggingProject}
})
