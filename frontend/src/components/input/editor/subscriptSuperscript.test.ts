import {describe, it, expect, beforeEach} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {ref} from 'vue'
import {Editor} from '@tiptap/core'

import {createEditorExtensions, type EditorExtensionDeps} from './editorExtensions'

const stubDeps: EditorExtensionDeps = {
	t: key => key,
	isEditing: ref(true),
	isEditEnabled: () => true,
	placeholder: '',
	contentHasChanged: ref(false),
	bubbleSave: () => {},
	getEditor: () => undefined,
	uploadCallback: undefined,
	uploadAndInsertFiles: () => {},
	loadedAttachments: ref({}),
	attachmentService: {} as never,
}

beforeEach(() => {
	setActivePinia(createPinia())
})

describe('Subscript and superscript support', () => {
	const createEditor = (content: string = '') => {
		return new Editor({
			extensions: createEditorExtensions(stubDeps),
			content,
		})
	}

	const pasteHtml = (editor: Editor, html: string, text: string) => {
		const event = new Event('paste', {bubbles: true, cancelable: true}) as ClipboardEvent
		Object.defineProperty(event, 'clipboardData', {
			value: {
				getData: (type: string) => {
					if (type === 'text/html') {
						return html
					}

					if (type === 'text/plain') {
						return text
					}

					return ''
				},
				items: [],
			},
		})

		editor.view.dom.dispatchEvent(event)
	}

	it('preserves subscript and superscript through the HTML round-trip', () => {
		const html = '<p>H<sub>2</sub>O and x<sup>2</sup></p>'
		const editor = createEditor(html)

		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})

	it('preserves subscript and superscript when pasting clipboard html', () => {
		const html = '<p>H<sub>2</sub>O and x<sup>2</sup></p>'
		const editor = createEditor('<p></p>')
		editor.commands.focus('end')

		pasteHtml(editor, html, 'H2O and x2')

		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})
})
