import {nextTick, watch} from 'vue'

type FocusTarget = () => HTMLElement | null | undefined

// Lost: dropped to <body> with a removed element, or still on what started the action.
function focusIsLost(trigger: Element | null) {
	const active = document.activeElement
	return !active || active === document.body || active === trigger
}

function focusFirst(targets: FocusTarget[]) {
	for (const target of targets) {
		const element = target()
		if (element) {
			element.focus()
			return
		}
	}
}

// Call when the action starts; the returned function never steals focus the user moved elsewhere.
export function captureFocus() {
	const trigger = document.activeElement
	return async (...targets: FocusTarget[]) => {
		await nextTick()
		if (focusIsLost(trigger)) {
			focusFirst(targets)
		}
	}
}

export function useFocusAfterRemoval(isShown: () => boolean, target: FocusTarget) {
	watch(isShown, async (shown, wasShown) => {
		if (!wasShown || shown) {
			return
		}
		await nextTick()
		if (focusIsLost(null)) {
			focusFirst([target])
		}
	})
}
