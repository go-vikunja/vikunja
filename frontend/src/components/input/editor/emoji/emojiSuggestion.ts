import {VueRenderer} from '@tiptap/vue-3'
import {computePosition, flip, shift, offset, autoUpdate} from '@floating-ui/dom'
import type {Editor, Range} from '@tiptap/core'
import type {EditorState} from '@tiptap/pm/state'

import EmojiList from './EmojiList.vue'
import {loadEmojis, filterEmojis, type EmojiEntry} from './emojiData'

interface SuggestionProps {
	editor: Editor
	range: Range
	query: string
	clientRect?: (() => DOMRect | null) | null
	items: EmojiEntry[]
	command: (item: EmojiEntry) => void
	event?: KeyboardEvent
}

const SHORTCODE_RE = /^[a-zA-Z0-9_]*$/

export default function emojiSuggestionSetup() {
	return {
		char: ':',
		allowedPrefixes: [' ', '\t', '\n'],
		startOfLine: false,

		allow: ({state, range}: {state: EditorState, range: Range}) => {
			const text = state.doc.textBetween(range.from, range.to, '\n', '\n')
			// Drop the leading ':' trigger character.
			const query = text.startsWith(':') ? text.slice(1) : text
			return SHORTCODE_RE.test(query)
		},

		items: async ({query}: {query: string}): Promise<EmojiEntry[]> => {
			if (query === '') return []
			try {
				const index = await loadEmojis()
				return filterEmojis(index, query)
			} catch (err) {
				console.error('Failed to load emoji index:', err)
				return []
			}
		},

		command: ({editor, range, props}: {editor: Editor, range: Range, props: EmojiEntry}) => {
			editor
				.chain()
				.focus()
				.deleteRange(range)
				.insertContent(props.emoji)
				.run()
		},

		render: () => {
			let component: VueRenderer
			let popupElement: HTMLElement | null = null
			let cleanupFloating: (() => void) | null = null

			const virtualReference = {
				getBoundingClientRect: () => ({
					width: 0, height: 0, x: 0, y: 0, top: 0, left: 0, right: 0, bottom: 0,
				} as DOMRect),
			}

			const mount = (props: SuggestionProps) => {
				component = new VueRenderer(EmojiList, {
					props,
					editor: props.editor,
				})
				if (!props.clientRect) return

				popupElement = document.createElement('div')
				popupElement.style.position = 'absolute'
				popupElement.style.top = '0'
				popupElement.style.left = '0'
				popupElement.style.zIndex = '4700'
				popupElement.appendChild(component.element!)
				document.body.appendChild(popupElement)

				const rect = props.clientRect()
				if (!rect) {
					unmount()
					return
				}
				virtualReference.getBoundingClientRect = () => rect

				const updatePosition = () => {
					computePosition(virtualReference, popupElement!, {
						placement: 'bottom-start',
						middleware: [offset(8), flip(), shift({padding: 8})],
					}).then(({x, y}) => {
						if (popupElement) {
							popupElement.style.left = `${x}px`
							popupElement.style.top = `${y}px`
						}
					})
				}
				updatePosition()
				cleanupFloating = autoUpdate(virtualReference, popupElement, updatePosition)
			}

			const unmount = () => {
				if (cleanupFloating) {
					cleanupFloating()
					cleanupFloating = null
				}
				if (popupElement) {
					document.body.removeChild(popupElement)
					popupElement = null
				}
				component?.destroy()
			}

			return {
				onStart: (props: SuggestionProps) => {
					if (!props.items.length && props.query === '') return
					mount(props)
				},

				onUpdate(props: SuggestionProps) {
					if (!popupElement) {
						if (props.items.length || props.query !== '') mount(props)
						return
					}
					component?.updateProps(props)
					if (!props.clientRect) return
					const rect = props.clientRect()
					if (rect) virtualReference.getBoundingClientRect = () => rect
				},

				onKeyDown(props: {event: KeyboardEvent}) {
					if (props.event.key === 'Escape') {
						if (props.event.isComposing) return false
						if (popupElement) popupElement.style.display = 'none'
						return true
					}
					return component?.ref?.onKeyDown(props)
				},

				onExit: unmount,
			}
		},
	}
}
