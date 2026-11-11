import {autoUpdate, computePosition, flip, offset, shift} from '@floating-ui/dom'

export interface SuggestionPopup {
	readonly element: HTMLElement
	reposition(): void
	destroy(): void
}

// autoUpdate ticks (visualViewport listeners, ResizeObserver) can land after teardown, so the popup
// owns its subscription and drops them instead of measuring an element that is already gone.
//
// getReferenceRect must stay live rather than being snapshotted: it reads the caret decoration on
// every tick, which is what keeps the popup glued to the text while the editor scrolls. contextElement
// puts the editor's scroll ancestors into autoUpdate's listener set so those ticks happen at all.
export function createSuggestionPopup(
	container: HTMLElement,
	content: Element,
	getReferenceRect: () => DOMRect | null,
	contextElement: Element,
): SuggestionPopup {
	const element = document.createElement('div')
	element.style.position = 'absolute'
	element.style.top = '0'
	element.style.left = '0'
	element.style.zIndex = '4700'
	element.appendChild(content)
	container.appendChild(element)

	let destroyed = false

	const reference = {
		getBoundingClientRect: () => getReferenceRect() ?? new DOMRect(),
		contextElement,
	}

	function updatePosition() {
		if (destroyed) {
			return
		}

		void computePosition(reference, element, {
			placement: 'bottom-start',
			middleware: [
				offset(8),
				flip(),
				shift({padding: 8}),
			],
		}).then(({x, y}) => {
			if (destroyed) {
				return
			}

			element.style.left = `${x}px`
			element.style.top = `${y}px`
		})
	}

	const stopAutoUpdate = autoUpdate(reference, element, updatePosition)

	return {
		element,

		reposition: updatePosition,

		destroy() {
			if (destroyed) {
				return
			}

			destroyed = true
			stopAutoUpdate()
			element.remove()
		},
	}
}
