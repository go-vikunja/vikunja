import { ref } from 'vue'

export interface GanttBarModel {
	id: string
	start: Date
	end: Date
	meta?: {
		label?: string
		color?: string
		hasActualDates?: boolean
		isDone?: boolean
		task?: unknown
	}
}
export interface UseGanttBarOptions {
	model: GanttBarModel
	timelineStart: Date
	timelineEnd: Date
	onUpdate?: (id: string, newStart: Date, newEnd: Date) => void
}

export function useGanttBar(options: UseGanttBarOptions) {
	const dragging = ref(false)
	const selected = ref(false)
	const focused = ref(false)

	function onFocus() {
		focused.value = true
	}

	function onBlur() {
		focused.value = false
	}

	function onKeyDown(e: KeyboardEvent) {
		if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {

			e.preventDefault()

			console.log('key')

			const dir = e.key === 'ArrowRight' ? 1 : -1
			const newStart = new Date(options.model.start)
			newStart.setDate(newStart.getDate() + dir)
			const newEnd = new Date(options.model.end)
			newEnd.setDate(newEnd.getDate() + dir)

			// Update the model for immediate visual feedback
			options.model.start = newStart
			options.model.end = newEnd

			// Notify parent component to persist the change
			if (options.onUpdate) {
				options.onUpdate(options.model.id, newStart, newEnd)
			}
		}
	}

	return {
		dragging,
		selected,
		focused,
		onFocus,
		onBlur,
		onKeyDown,
	}
}
