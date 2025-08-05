import {createRandomID} from '@/helpers/randomId'
import {computePosition, flip, shift, offset} from '@floating-ui/dom'
import {nextTick} from 'vue'
import {eventToHotkeyString} from '@github/hotkey'

export default function inputPrompt(pos: ClientRect, oldValue: string = ''): Promise<string> {
	return new Promise((resolve) => {
		const id = 'link-input-' + createRandomID()

		// Create popup element
		const popupElement = document.createElement('div')
		popupElement.style.position = 'absolute'
		popupElement.style.top = '0'
		popupElement.style.left = '0'
		popupElement.style.zIndex = '1000'
		popupElement.style.background = 'white'
		popupElement.style.border = '1px solid #ccc'
		popupElement.style.borderRadius = '4px'
		popupElement.style.padding = '8px'
		popupElement.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)'
		popupElement.innerHTML = `<div><input class="input" placeholder="URL" id="${id}" value="${oldValue}"/></div>`
		document.body.appendChild(popupElement)

		// Virtual reference for positioning
		const virtualReference = {
			getBoundingClientRect: () => pos,
		}

		// Position the popup
		computePosition(virtualReference, popupElement, {
			placement: 'top-start',
			middleware: [
				offset(8),
				flip(),
				shift({ padding: 8 }),
			],
		}).then(({ x, y }) => {
			popupElement.style.left = `${x}px`
			popupElement.style.top = `${y}px`
		})

		nextTick(() => document.getElementById(id)?.focus())

		const cleanup = () => {
			if (document.body.contains(popupElement)) {
				document.body.removeChild(popupElement)
			}
		}

		document.getElementById(id)?.addEventListener('keydown', event => {
			const hotkeyString = eventToHotkeyString(event)
			if (hotkeyString !== 'Enter') {
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
