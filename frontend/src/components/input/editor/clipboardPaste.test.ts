import {describe, it, expect, afterEach, beforeEach} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {ref} from 'vue'
import {Editor} from '@tiptap/core'
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

function pasteHTML(editor: Editor, html: string, at = 1) {
	editor.view.dispatch(editor.state.tr.setSelection(TextSelection.create(editor.state.doc, at)))
	editor.view.pasteHTML(html)
}

function sliceContext(context: unknown[]) {
	return JSON.stringify(context).replaceAll('"', '&quot;')
}

beforeEach(() => {
	setActivePinia(createPinia())
})

afterEach(() => {
	document.body.innerHTML = ''
})

describe('pasting html the schema cannot hold', () => {
	// A ProseMirror editor whose list items hold inline content directly puts a context
	// on the clipboard that makes our listItem swallow bare text, see FRONTEND-OSS-2H9.
	it.each([
		['a list item holding text', ['bulletList', null, 'listItem', null], '<ul><li><p>pasted</p></li></ul>'],
		['an ordered list item holding text', ['orderedList', null, 'listItem', null], '<ol><li><p>pasted</p></li></ol>'],
		['a task item holding text', ['taskList', null, 'taskItem', null], '<p>pasted</p>'],
	])('wraps text pasted into %s', (_name, context, expected) => {
		const editor = createEditor('<p></p>')

		expect(() => pasteHTML(editor, `<span data-pm-slice="0 0 ${sliceContext(context)}">pasted</span>`)).not.toThrow()
		expect(() => editor.state.doc.check()).not.toThrow()
		expect(editor.getHTML()).toContain(expected)
	})

	// Copying a list item from the nested list inside it yields a listItem whose first
	// child is that nested list instead of the paragraph the schema requires.
	it.each([
		['an empty document', '<p></p>', 1],
		['a list item', '<ol><li><p>hello</p></li></ol>', 4],
	])('fills in the missing paragraph when pasting a nested list into %s', (_name, content, at) => {
		const editor = createEditor(content)
		const html = '<li data-pm-slice="1 0 []"><ol><li><p>nested</p></li></ol></li>'

		expect(() => pasteHTML(editor, html, at)).not.toThrow()
		expect(() => editor.state.doc.check()).not.toThrow()
		expect(editor.getHTML()).toContain('<li><p>nested</p></li>')
	})

	it('leaves valid pasted content alone', () => {
		const editor = createEditor('<p></p>')

		pasteHTML(editor, '<ol><li><p>one</p></li><li><p>two</p></li></ol>')

		expect(editor.getHTML()).toBe('<ol><li><p>one</p></li><li><p>two</p></li></ol><p></p>')
	})

	it('keeps a list copied out of the editor intact', () => {
		const source = createEditor('<ol><li><p>one</p></li><li><p>two</p></li></ol>')
		const copied = source.view.serializeForClipboard(source.state.doc.slice(0, source.state.doc.content.size)).dom as HTMLElement

		const editor = createEditor('<p>hello</p>')
		pasteHTML(editor, copied.innerHTML, 6)

		expect(editor.getHTML()).toBe('<p>hello</p><ol><li><p>one</p></li><li><p>two</p></li></ol><p></p>')
	})
})
