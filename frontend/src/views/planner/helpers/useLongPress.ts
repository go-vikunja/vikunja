import {onBeforeUnmount} from 'vue'

const MOVE_THRESHOLD_PX = 10
const LONG_PRESS_MS = 500

// Touch/pen long-press-to-create: fires the callback after a still press and
// bails out if the pointer moves first, so the gesture doesn't hijack
// scrolling. Mouse pointers are ignored — they have click/dblclick paths.
// Listeners are torn down on unmount so a mid-press re-render can't leak them
// onto document.
export function useLongPress() {
	let timer: ReturnType<typeof setTimeout> | undefined
	let move: ((e: PointerEvent) => void) | null = null
	let end: (() => void) | null = null

	function detach() {
		clearTimeout(timer)
		if (move) {
			document.removeEventListener('pointermove', move)
		}
		if (end) {
			document.removeEventListener('pointerup', end)
			document.removeEventListener('pointercancel', end)
		}
		move = null
		end = null
	}

	function start(event: PointerEvent, onTrigger: () => void) {
		if (event.pointerType === 'mouse') {
			return
		}
		detach()
		const startX = event.clientX
		const startY = event.clientY
		let moved = false
		move = (e: PointerEvent) => {
			if (Math.abs(e.clientX - startX) > MOVE_THRESHOLD_PX || Math.abs(e.clientY - startY) > MOVE_THRESHOLD_PX) {
				moved = true
				detach()
			}
		}
		end = detach
		document.addEventListener('pointermove', move)
		document.addEventListener('pointerup', end)
		document.addEventListener('pointercancel', end)
		timer = setTimeout(() => {
			detach()
			if (!moved) {
				onTrigger()
			}
		}, LONG_PRESS_MS)
	}

	onBeforeUnmount(detach)

	return {start, detach}
}
