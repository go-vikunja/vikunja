import {VueRenderer} from '@tiptap/vue-3'
import {computePosition, flip, shift, offset, autoUpdate} from '@floating-ui/dom'

import CommandsList from './CommandsList.vue'

export default function suggestionSetup(t) {
	return {
		items: ({query}: { query: string }) => {
			return [
				{
					title: t('input.editor.text'),
					description: t('input.editor.textTooltip'),
					icon: 'fa-font',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.setNode('paragraph', {level: 1})
							.run()
					},
				},
				{
					title: t('input.editor.heading1'),
					description: t('input.editor.heading1Tooltip'),
					icon: 'fa-header',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.setNode('heading', {level: 1})
							.run()
					},
				},
				{
					title: t('input.editor.heading2'),
					description: t('input.editor.heading2Tooltip'),
					icon: 'fa-header',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.setNode('heading', {level: 2})
							.run()
					},
				},
				{
					title: t('input.editor.heading3'),
					description: t('input.editor.heading3Tooltip'),
					icon: 'fa-header',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.setNode('heading', {level: 2})
							.run()
					},
				},
				{
					title: t('input.editor.bulletList'),
					description: t('input.editor.bulletListTooltip'),
					icon: 'fa-list-ul',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.toggleBulletList()
							.run()
					},
				},
				{
					title: t('input.editor.orderedList'),
					description: t('input.editor.orderedListTooltip'),
					icon: 'fa-list-ol',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.toggleOrderedList()
							.run()
					},
				},
				{
					title: t('input.editor.taskList'),
					description: t('input.editor.taskListTooltip'),
					icon: 'fa-list-check',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.toggleTaskList()
							.run()
					},
				},
				{
					title: t('input.editor.quote'),
					description: t('input.editor.quoteTooltip'),
					icon: 'fa-quote-right',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.toggleBlockquote()
							.run()
					},
				},
				{
					title: t('input.editor.code'),
					description: t('input.editor.codeTooltip'),
					icon: 'fa-code',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.toggleCodeBlock()
							.run()
					},
				},
				{
					title: t('input.editor.image'),
					description: t('input.editor.imageTooltip'),
					icon: 'fa-image',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.run()
						document.getElementById('tiptap__image-upload').click()
					},
				},
				{
					title: t('input.editor.horizontalRule'),
					description: t('input.editor.horizontalRuleTooltip'),
					icon: 'fa-ruler-horizontal',
					command: ({editor, range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.setHorizontalRule()
							.run()
					},
				},
			].filter(item => item.title.toLowerCase().startsWith(query.toLowerCase()))
		},

		render: () => {
			let component: VueRenderer
			let popupElement: HTMLElement | null = null
			let cleanupFloating: (() => void) | null = null

			const virtualReference = {
				getBoundingClientRect: () => ({
					width: 0,
					height: 0,
					x: 0,
					y: 0,
					top: 0,
					left: 0,
					right: 0,
					bottom: 0,
				} as DOMRect),
			}

			return {
				onStart: props => {
					component = new VueRenderer(CommandsList, {
						// using vue 2:
						// parent: this,
						// propsData: props,
						props,
						editor: props.editor,
					})

					if (!props.clientRect) {
						return
					}

					// Create popup element
					popupElement = document.createElement('div')
					popupElement.style.position = 'absolute'
					popupElement.style.top = '0'
					popupElement.style.left = '0'
					popupElement.style.zIndex = '1000'
					popupElement.appendChild(component.element!)
					document.body.appendChild(popupElement)

					// Update virtual reference
					const rect = props.clientRect()
					virtualReference.getBoundingClientRect = () => rect

					// Set up floating positioning
					const updatePosition = () => {
						computePosition(virtualReference, popupElement!, {
							placement: 'bottom-start',
							middleware: [
								offset(8),
								flip(),
								shift({ padding: 8 }),
							],
						}).then(({ x, y }) => {
							if (popupElement) {
								popupElement.style.left = `${x}px`
								popupElement.style.top = `${y}px`
							}
						})
					}

					updatePosition()
					cleanupFloating = autoUpdate(virtualReference, popupElement, updatePosition)
				},

				onUpdate(props) {
					component.updateProps(props)

					if (!props.clientRect || !popupElement) {
						return
					}

					// Update virtual reference
					const rect = props.clientRect()
					virtualReference.getBoundingClientRect = () => rect
				},

				onKeyDown(props) {
					if (props.event.key === 'Escape') {
						if (popupElement) {
							popupElement.style.display = 'none'
						}

						return true
					}

					return component.ref?.onKeyDown(props)
				},

				onExit() {
					if (cleanupFloating) {
						cleanupFloating()
					}
					if (popupElement) {
						document.body.removeChild(popupElement)
						popupElement = null
					}
					component.destroy()
				},
			}
		},
	}
}
