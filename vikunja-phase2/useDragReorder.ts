import {ref} from 'vue'

/**
 * Composable for HTML5 drag-to-reorder on arrays.
 * Usage:
 *   const {dragIndex, dragOverIndex, onDragStart, onDragOver, onDragEnd, onDrop} = useDragReorder(myArray)
 *
 * Template:
 *   <div
 *     v-for="(item, i) in myArray"
 *     draggable="true"
 *     :class="{ 'is-dragging': dragIndex === i, 'is-drag-over': dragOverIndex === i }"
 *     @dragstart="onDragStart(i, $event)"
 *     @dragover.prevent="onDragOver(i)"
 *     @dragleave="onDragLeave()"
 *     @drop.prevent="onDrop(i)"
 *     @dragend="onDragEnd"
 *   >
 */
export function useDragReorder<T>(items: { value: T[] }) {
	const dragIndex = ref<number | null>(null)
	const dragOverIndex = ref<number | null>(null)

	function onDragStart(index: number, event: DragEvent) {
		dragIndex.value = index
		if (event.dataTransfer) {
			event.dataTransfer.effectAllowed = 'move'
			// Required for Firefox
			event.dataTransfer.setData('text/plain', String(index))
		}
	}

	function onDragOver(index: number) {
		dragOverIndex.value = index
	}

	function onDragLeave() {
		dragOverIndex.value = null
	}

	function onDrop(targetIndex: number) {
		if (dragIndex.value === null || dragIndex.value === targetIndex) {
			dragOverIndex.value = null
			return
		}

		const arr = items.value
		const [moved] = arr.splice(dragIndex.value, 1)
		arr.splice(targetIndex, 0, moved)

		dragIndex.value = null
		dragOverIndex.value = null
	}

	function onDragEnd() {
		dragIndex.value = null
		dragOverIndex.value = null
	}

	return {
		dragIndex,
		dragOverIndex,
		onDragStart,
		onDragOver,
		onDragLeave,
		onDrop,
		onDragEnd,
	}
}
