import {ref} from 'vue'

export interface GanttBarModel { 
	id: string; 
	start: Date; 
	end: Date;
	meta?: {
		label?: string;
		color?: string;
		hasActualDates?: boolean;
		isDone?: boolean;
		task?: unknown;
	}
}
export interface UseGanttBarOptions {
  model: GanttBarModel
  timelineStart: Date
  timelineEnd: Date
}

export function useGanttBar(options: UseGanttBarOptions) {
	const dragging = ref(false)
	const selected = ref(false)
	const focused = ref(false)

	function onPointerDown(e: PointerEvent) {
		dragging.value = true
		const startX = e.clientX
		const handleMove = (evt: PointerEvent) => {
			const diff = evt.clientX - startX
			// Use 30px per day to match our styling
			const days = Math.round(diff / 30)
			const newStart = new Date(options.model.start)
			newStart.setDate(newStart.getDate() + days)
			const newEnd = new Date(options.model.end)
			newEnd.setDate(newEnd.getDate() + days)
		}
		const stop = () => {
			dragging.value = false
			window.removeEventListener('pointermove', handleMove)
			window.removeEventListener('pointerup', stop)
		}
		window.addEventListener('pointermove', handleMove)
		window.addEventListener('pointerup', stop)
	}

	function onFocus() { focused.value = true }
	function onBlur() { focused.value = false }
	function onKeyDown(e: KeyboardEvent) {
		if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
			const dir = e.key === 'ArrowRight' ? 1 : -1
			const newStart = new Date(options.model.start)
			newStart.setDate(newStart.getDate() + dir)
			const newEnd = new Date(options.model.end)
			newEnd.setDate(newEnd.getDate() + dir)
		}
	}

	return { dragging, selected, focused, onPointerDown, onFocus, onBlur, onKeyDown }
}
