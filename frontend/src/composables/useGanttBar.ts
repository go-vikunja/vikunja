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

	function changeSize(direction: 'left' | 'right', modifier: -1 | 1) {
		const newStart = new Date(options.model.start)
		const newEnd = new Date(options.model.end)

		if (direction === 'left') {
			// Shift+Left: Expand task to the left (move start date earlier)
			newStart.setDate(newStart.getDate() - 1 * modifier)
		} else {
			// Shift+Right: Expand task to the right (move end date later)  
			newEnd.setDate(newEnd.getDate() + 1 * modifier)
		}

		// Validate that start is before end (maintain minimum 1 day duration)
		if (newStart < newEnd) {
			options.model.start = newStart
			options.model.end = newEnd

			if (options.onUpdate) {
				options.onUpdate(options.model.id, newStart, newEnd)
			}
		}
	}

	function onKeyDown(e: KeyboardEvent) {
		// task expanding
		if (e.shiftKey) {
			if (e.key === 'ArrowLeft') {
				e.preventDefault()
				changeSize('left', 1)
			}
			if (e.key === 'ArrowRight') {
				e.preventDefault()
				changeSize('right', 1)
			}
		}
		// task shrinking
		else if (e.ctrlKey) {
			if (e.key === 'ArrowLeft') {
				e.preventDefault()
				changeSize('left', -1)
			}
			if (e.key === 'ArrowRight') {
				e.preventDefault()
				changeSize('right', -1)
			}
		}
		// task movement
		else if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
			e.preventDefault()

			const dir = e.key === 'ArrowRight' ? 1 : -1
			const newStart = new Date(options.model.start)
			newStart.setDate(newStart.getDate() + dir)
			const newEnd = new Date(options.model.end)
			newEnd.setDate(newEnd.getDate() + dir)

			options.model.start = newStart
			options.model.end = newEnd

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
