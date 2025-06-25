import {ref} from 'vue'

export interface GanttBarModel { id: string; start: Date; end: Date }
export interface UseGanttBarOptions {
  model: GanttBarModel
  timelineStart: Date
  timelineEnd: Date
  onMove(id: string, newStart: Date, newEnd: Date): void
  window?: Window | undefined
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
			// simplified: move one day per 20px
			const days = Math.round(diff / 20)
			const newStart = new Date(options.model.start)
			newStart.setDate(newStart.getDate() + days)
			const newEnd = new Date(options.model.end)
			newEnd.setDate(newEnd.getDate() + days)
			options.onMove(options.model.id, newStart, newEnd)
		}
		const stop = () => {
			dragging.value = false
			opts.removeEventListener('pointermove', handleMove)
			opts.removeEventListener('pointerup', stop)
		}
		const opts = options.window ?? window
		opts.addEventListener('pointermove', handleMove)
		opts.addEventListener('pointerup', stop)
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
			options.onMove(options.model.id, newStart, newEnd)
		}
	}

	return { dragging, selected, focused, onPointerDown, onFocus, onBlur, onKeyDown }
}
