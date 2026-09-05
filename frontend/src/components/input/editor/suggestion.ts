import type {Editor, Range} from '@tiptap/core'
import {VueRenderer} from '@tiptap/vue-3'

import CommandsList from './CommandsList.vue'
import {getPopupContainer} from './popupContainer'
import {createSuggestionPopup, type SuggestionPopup} from './suggestionPopup'

type TranslateFunction = (key: string) => string

interface SuggestionProps {
	editor: Editor
	clientRect?: () => DOMRect
	command: (item: {command: (params: {editor: Editor, range: Range}) => void}) => void
	items: unknown[]
	event?: KeyboardEvent
}

export default function suggestionSetup(t: TranslateFunction) {
	return {
		items: ({query}: { query: string }) => {
			return [
				{
					title: t('input.editor.text'),
					description: t('input.editor.textTooltip'),
					icon: 'fa-font',
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
					command: ({editor, range}: {editor: Editor, range: Range}) => {
						editor
							.chain()
							.focus()
							.deleteRange(range)
							.run()
						const uploadElement = document.getElementById('tiptap__image-upload')
						if (uploadElement) {
							uploadElement.click()
						}
					},
				},
				{
					title: t('input.editor.horizontalRule'),
					description: t('input.editor.horizontalRuleTooltip'),
					icon: 'fa-ruler-horizontal',
					command: ({editor, range}: {editor: Editor, range: Range}) => {
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
			let popup: SuggestionPopup | null = null

			return {
				onStart: (props: SuggestionProps) => {
					component = new VueRenderer(CommandsList, {
						// using vue 2:
						// parent: this,
						// propsData: props,
						props,
						editor: props.editor,
					})

					const rect = props.clientRect?.()
					if (!rect) {
						return
					}

					popup = createSuggestionPopup(getPopupContainer(props.editor), component.element!, rect)
				},

				onUpdate(props: SuggestionProps) {
					component.updateProps(props)

					const rect = props.clientRect?.()
					if (rect) {
						popup?.setReferenceRect(rect)
					}
				},

				onKeyDown(props: SuggestionProps) {
					if (props.event && props.event.key === 'Escape') {
						if (props.event.isComposing) {
							return false
						}

						if (popup) {
							popup.element.style.display = 'none'
						}

						return true
					}

					return component.ref?.onKeyDown(props)
				},

				onExit() {
					popup?.destroy()
					popup = null
					component.destroy()
				},
			}
		},
	}
}
