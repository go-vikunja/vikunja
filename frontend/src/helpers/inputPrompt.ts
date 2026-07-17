import {createRandomID} from '@/helpers/randomId'
import {computePosition, flip, shift, offset} from '@floating-ui/dom'
import {nextTick} from 'vue'
import {eventToShortcutString} from '@/helpers/shortcut'
import type {Editor} from '@tiptap/core'
import {getPopupContainer} from '@/components/input/editor/popupContainer'

export default function inputPrompt(pos: ClientRect, oldValue: string = '', editor?: Editor, placeholder: string = 'URL'): Promise<string | null> {
	return new Promise((resolve) => {
		const id = 'link-input-' + createRandomID()
		// Append inside the open task <dialog> (top-layer) when present, otherwise
		// document.body. A body-level popup is painted behind a showModal() dialog
		// and unfocusable through its focus trap, breaking the link prompt in the
		// Kanban task popup (#2940).
		const container = getPopupContainer(editor)

		// Create popup element
		const popupElement = document.createElement('div')
		popupElement.style.position = 'fixed'
		popupElement.style.top = '0'
		popupElement.style.left = '0'
		popupElement.style.zIndex = '4700'
		popupElement.style.background = 'var(--white)'
		popupElement.style.border = '1px solid var(--grey-300)'
		popupElement.style.borderRadius = '4px'
		popupElement.style.padding = '8px'
		popupElement.style.boxShadow = 'var(--shadow-md)'
		const wrapperDiv = document.createElement('div')
		const inputElement = document.createElement('input')
		inputElement.className = 'input'
		inputElement.placeholder = placeholder
		inputElement.setAttribute('aria-label', placeholder)
		inputElement.id = id
		inputElement.value = oldValue
		wrapperDiv.appendChild(inputElement)
		popupElement.appendChild(wrapperDiv)
		container.appendChild(popupElement)

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

		nextTick(() => {
			const inputEl = document.getElementById(id) as HTMLInputElement | null
			inputEl?.focus()

			// Over a selected image (a NodeSelection) the editor reclaims DOM focus
			// from the input on a browser-internal timer; re-assert it the one time.
			inputEl?.addEventListener('blur', () => {
				if (editor && document.activeElement === editor.view.dom) {
					inputEl.focus()
				}
			}, {once: true})
		})

		// The prompt is a sub-modal of the enclosing task <dialog>. Native modal
		// dialogs close themselves on Escape ("cancel"); swallow that while the
		// prompt is open so Escape only dismisses the prompt, not the task dialog.
		const dialog = container.closest('dialog') as HTMLDialogElement | null
		const handleDialogCancel = (event: Event) => event.preventDefault()
		dialog?.addEventListener('cancel', handleDialogCancel)

		const handleClickOutside = (event: MouseEvent) => {
			if (!popupElement.contains(event.target as Node)) {
				resolve(null)
				cleanup()
			}
		}

		const cleanup = () => {
			window.removeEventListener('scroll', handleScroll, true)
			document.removeEventListener('click', handleClickOutside)
			dialog?.removeEventListener('cancel', handleDialogCancel)
			if (container.contains(popupElement)) {
				container.removeChild(popupElement)
			}
		}

		document.getElementById(id)?.addEventListener('keydown', event => {
			const shortcutString = eventToShortcutString(event)

			if (shortcutString === 'Escape') {
				// Stop the native <dialog> from closing on Escape; cancel the prompt only.
				event.preventDefault()
				event.stopPropagation()
				resolve(null)
				cleanup()
				return
			}

			if (shortcutString !== 'Enter') {
				return
			}

			if (event.isComposing) {
				return
			}

			const url = (event.target as HTMLInputElement).value

			resolve(url)
			cleanup()
		})

		// Add slight delay to prevent immediate closing
		setTimeout(() => {
			document.addEventListener('click', handleClickOutside)
		}, 100)

	})
}
