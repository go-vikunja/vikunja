import {autoUpdate, computePosition, flip, offset, shift} from '@floating-ui/dom'

export interface SuggestionPopup {
	readonly element: HTMLElement
	setReferenceRect(rect: DOMRect): void
	destroy(): void
}

// autoUpdate ticks (visualViewport listeners, ResizeObserver) can land after teardown, so the popup
// owns its subscription and drops them instead of measuring an element that is already gone.
export function createSuggestionPopup(container: HTMLElement, content: Element, rect: DOMRect): SuggestionPopup {
	const element = document.createElement('div')
	element.style.position = 'absolute'
	element.style.top = '0'
	element.style.left = '0'
	element.style.zIndex = '4700'
	element.appendChild(content)
	container.appendChild(element)

	let referenceRect = rect
	let destroyed = false

	const reference = {
		getBoundingClientRect: () => referenceRect,
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

		setReferenceRect(next: DOMRect) {
			referenceRect = next
		},

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
