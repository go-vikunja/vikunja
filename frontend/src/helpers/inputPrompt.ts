import {createRandomID} from '@/helpers/randomId'
import {computePosition, flip, shift, offset} from '@floating-ui/dom'
import {nextTick} from 'vue'
import {eventToHotkeyString} from '@github/hotkey'

export default function inputPrompt(pos: ClientRect, oldValue: string = ''): Promise<string> {
	return new Promise((resolve) => {
		const id = 'link-input-' + createRandomID()

		// Create popup element
		const popupElement = document.createElement('div')
		popupElement.style.position = 'fixed'
		popupElement.style.top = '0'
		popupElement.style.left = '0'
		popupElement.style.zIndex = '4700'
		popupElement.style.background = 'white'
		popupElement.style.border = '1px solid #ccc'
		popupElement.style.borderRadius = '4px'
		popupElement.style.padding = '8px'
		popupElement.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)'
		popupElement.innerHTML = `<div><input class="input" placeholder="URL" id="${id}" value="${oldValue}"/></div>`
		document.body.appendChild(popupElement)

		// Create a local mutable copy of the position for scroll tracking
		let currentRect = new DOMRect(pos.left, pos.top, pos.width, pos.height)

		// Virtual reference for positioning
		const virtualReference = {
			getBoundingClientRect: () => currentRect,
		}

		// Function to update popup position
		const updatePosition = () => {
			computePosition(virtualReference, popupElement, {
				placement: 'top-start',
				strategy: 'fixed',
				middleware: [
					offset(8),
					flip(),
					shift({ padding: 8 }),
				],
			}).then(({ x, y }) => {
				popupElement.style.left = `${x}px`
				popupElement.style.top = `${y}px`
			})
		}

		// Position the popup initially
		updatePosition()

		// Track scroll position
		let lastScrollY = window.scrollY
		let lastScrollX = window.scrollX

		// Update position on scroll
		const handleScroll = () => {
			const deltaY = window.scrollY - lastScrollY
			const deltaX = window.scrollX - lastScrollX

			// Update the local mutable rect to account for scroll
			currentRect = new DOMRect(
				currentRect.x - deltaX,
				currentRect.y - deltaY,
				currentRect.width,
				currentRect.height,
			)

			lastScrollY = window.scrollY
			lastScrollX = window.scrollX

			updatePosition()
		}

		window.addEventListener('scroll', handleScroll, true)

		nextTick(() => document.getElementById(id)?.focus())

		const cleanup = () => {
			window.removeEventListener('scroll', handleScroll, true)
			if (document.body.contains(popupElement)) {
				document.body.removeChild(popupElement)
			}
		}

		document.getElementById(id)?.addEventListener('keydown', event => {
			const hotkeyString = eventToHotkeyString(event)
			if (hotkeyString !== 'Enter') {
				return
			}

			if (event.isComposing) {
				return
			}

			const url = (event.target as HTMLInputElement).value

			resolve(url)
			cleanup()
		})

		// Close on click outside
		const handleClickOutside = (event: MouseEvent) => {
			if (!popupElement.contains(event.target as Node)) {
				resolve('')
				cleanup()
				document.removeEventListener('click', handleClickOutside)
			}
		}

		// Add slight delay to prevent immediate closing
		setTimeout(() => {
			document.addEventListener('click', handleClickOutside)
		}, 100)

	})
}
