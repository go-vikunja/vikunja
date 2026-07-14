import {computed, ref} from 'vue'
import {defineStore} from 'pinia'

export const useBulkTaskSelection = defineStore('bulkTaskSelection', () => {
	const selectedTaskIds = ref<number[]>([])
	const lastSelectedTaskId = ref<number | null>(null)

	const selectedCount = computed(() => selectedTaskIds.value.length)
	const hasSelection = computed(() => selectedCount.value > 0)

	function isSelected(taskId: number): boolean {
		return selectedTaskIds.value.includes(taskId)
	}

	function toggle(taskId: number) {
		if (isSelected(taskId)) {
			selectedTaskIds.value = selectedTaskIds.value.filter(id => id !== taskId)
		} else {
			selectedTaskIds.value = [...selectedTaskIds.value, taskId]
		}

		lastSelectedTaskId.value = taskId
	}

	function toggleRange(visibleTaskIds: number[], taskId: number, shiftKey: boolean) {
		if (!shiftKey || lastSelectedTaskId.value === null) {
			toggle(taskId)
			return
		}

		const startIndex = visibleTaskIds.indexOf(lastSelectedTaskId.value)
		const endIndex = visibleTaskIds.indexOf(taskId)

		if (startIndex === -1 || endIndex === -1) {
			toggle(taskId)
			return
		}

		const [start, end] = startIndex < endIndex
			? [startIndex, endIndex]
			: [endIndex, startIndex]

		const rangeIds = visibleTaskIds.slice(start, end + 1)

		selectedTaskIds.value = Array.from(new Set([
			...selectedTaskIds.value,
			...rangeIds,
		]))

		lastSelectedTaskId.value = taskId
	}

	function select(taskId: number) {
		if (!isSelected(taskId)) {
			selectedTaskIds.value = [...selectedTaskIds.value, taskId]
		}

		lastSelectedTaskId.value = taskId
	}

	function deselect(taskId: number) {
		selectedTaskIds.value = selectedTaskIds.value.filter(id => id !== taskId)

		if (lastSelectedTaskId.value === taskId) {
			lastSelectedTaskId.value = null
		}
	}

	function replace(taskIds: number[]) {
		selectedTaskIds.value = Array.from(new Set(taskIds))
		lastSelectedTaskId.value = taskIds.at(-1) ?? null
	}

	function selectMany(taskIds: number[]) {
		selectedTaskIds.value = Array.from(new Set([
			...selectedTaskIds.value,
			...taskIds,
		]))

		if (taskIds.length > 0) {
			lastSelectedTaskId.value = taskIds.at(-1) ?? null
		}
	}

	function clear() {
		selectedTaskIds.value = []
		lastSelectedTaskId.value = null
	}

	return {
		selectedTaskIds,
		selectedCount,
		hasSelection,
		lastSelectedTaskId,
		isSelected,
		toggle,
		toggleRange,
		select,
		deselect,
		replace,
		selectMany,
		clear,
	}
})