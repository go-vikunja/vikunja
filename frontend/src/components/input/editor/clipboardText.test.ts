import {describe, it, expect, afterEach, beforeEach} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {ref} from 'vue'
import {Editor} from '@tiptap/core'
import {TextSelection} from '@tiptap/pm/state'

import {createEditorExtensions, type EditorExtensionDeps} from './editorExtensions'

window.API_URL = 'http://localhost:3456'

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

function copyText(content: string, from?: number, to?: number) {
	const editor = createEditor(content)
	const {doc} = editor.state
	const selection = TextSelection.create(doc, from ?? 0, to ?? doc.content.size)
	return editor.view.serializeForClipboard(selection.content()).text
}

beforeEach(() => {
	setActivePinia(createPinia())
})

afterEach(() => {
	document.body.innerHTML = ''
})

describe('copying as plain text', () => {
	it('keeps list markers without extra blank lines', () => {
		expect(copyText('<p>start</p><ul><li><p>this</p></li><li><p>is</p></li></ul><p>end</p>'))
			.toBe('start\n\n- this\n- is\n\nend')
	})

	it('numbers ordered lists from their start', () => {
		expect(copyText('<ol start="3"><li><p>a</p></li><li><p>b</p></li></ol>')).toBe('3. a\n4. b')
	})

	it('indents nested lists', () => {
		expect(copyText('<ul><li><p>a</p><ul><li><p>b</p></li></ul></li></ul>')).toBe('- a\n  - b')
	})

	it('marks task items', () => {
		expect(copyText('<ul data-type="taskList"><li data-type="taskItem" data-checked="true"><p>done</p></li><li data-type="taskItem" data-checked="false"><p>todo</p></li></ul>'))
			.toBe('- [x] done\n- [ ] todo')
	})

	it('serializes headings, quotes, code and marks', () => {
		expect(copyText('<h2>Title</h2><blockquote><p>quoted</p></blockquote><pre><code>x = 1</code></pre><p><strong>bold</strong> <a href="https://example.com">link</a></p>'))
			.toBe('## Title\n\n> quoted\n\n```\nx = 1\n```\n\n**bold** [link](https://example.com)')
	})

	it('serializes tables', () => {
		expect(copyText('<table><tr><th><p>a</p></th><th><p>b</p></th></tr><tr><td><p>1</p></td><td><p>2</p></td></tr></table>'))
			.toBe('| a | b |\n| --- | --- |\n| 1 | 2 |')
	})

	it('copies a partial selection inside a paragraph as its text', () => {
		expect(copyText('<p>hello world</p>', 7, 12)).toBe('world')
	})
})
