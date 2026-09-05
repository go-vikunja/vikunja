import {describe, it, expect, afterEach, beforeEach} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {ref} from 'vue'
import {Editor} from '@tiptap/core'
import {Slice} from '@tiptap/pm/model'
import {TextSelection} from '@tiptap/pm/state'

import {createEditorExtensions, type EditorExtensionDeps} from './editorExtensions'

window.API_URL = 'http://localhost:3456/api/v1'

function createEditor(content: string) {
	const holder = document.createElement('div')
	document.body.appendChild(holder)

	let editor: Editor
	const deps: EditorExtensionDeps = {
		t: key => key,
		isEditing: ref(true),
		isEditEnabled: () => true,
		placeholder: '',
		contentHasChanged: ref(false),
		bubbleSave: () => {},
		getEditor: () => editor,
		uploadCallback: undefined,
		uploadAndInsertFiles: () => {},
	}

	editor = new Editor({element: holder, extensions: createEditorExtensions(deps), content})

	return editor
}

function paste(editor: Editor, text: string) {
	const event = {
		clipboardData: {
			items: [],
			getData: (type: string) => type === 'text/plain' ? text : '',
		},
	} as unknown as ClipboardEvent

	editor.view.someProp('handlePaste', handler => handler(editor.view, event, Slice.empty))
}

beforeEach(() => {
	setActivePinia(createPinia())
})

afterEach(() => {
	document.body.innerHTML = ''
})

describe('pasting markdown', () => {
	// list items the markdown parser leaves without a paragraph, see FRONTEND-OSS-2JY
	it.each([
		['an empty list item', '- ', '<ul><li><p></p></li></ul>'],
		['an empty list item after a paragraph', 'foo\n\n- \n', '<p>foo</p><ul><li><p></p></li></ul>'],
		['a list item starting with a nested list', '- \n  - x', '<ul><li><p></p><ul><li><p>x</p></li></ul></li></ul>'],
		['a list item starting with a heading', '- foo\n  - \n', '<ul><li><p></p><h2>foo</h2></li></ul>'],
	])('inserts %s', (_name, text, expected) => {
		const editor = createEditor('<p></p>')

		expect(() => paste(editor, text)).not.toThrow()
		expect(() => editor.state.doc.check()).not.toThrow()
		expect(editor.getHTML()).toContain(expected)
	})

	it('converts markdown to nodes', () => {
		const editor = createEditor('<p></p>')

		paste(editor, '# Heading\n\ntext with *em*\n\n- a\n- b\n')

		expect(editor.getHTML()).toBe('<h1>Heading</h1><p>text with <em>em</em></p><ul><li><p>a</p></li><li><p>b</p></li></ul><p></p>')
	})

	it('inserts markdown at the cursor', () => {
		const editor = createEditor('<p>hello world</p>')
		editor.view.dispatch(editor.state.tr.setSelection(TextSelection.create(editor.state.doc, 7)))

		paste(editor, '- a\n- b\n')

		expect(editor.getHTML()).toBe('<p>hello </p><ul><li><p>a</p></li><li><p>b</p></li></ul><p>world</p>')
	})
})
