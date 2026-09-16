import {describe, it, expect, afterEach, beforeEach, vi} from 'vitest'
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
		loadedAttachments: ref({}),
		attachmentService: {getBlobUrl: vi.fn()} as never,
	}

	editor = new Editor({element: holder, extensions: createEditorExtensions(deps), content})

	return editor
}

function select(editor: Editor, from: number, to: number) {
	editor.view.dispatch(editor.state.tr.setSelection(
		TextSelection.create(editor.state.doc, from, to),
	))
}

function pressEnter(editor: Editor) {
	editor.view.someProp('handleKeyDown', handler => handler(
		editor.view,
		new KeyboardEvent('keydown', {key: 'Enter', keyCode: 13}),
	))
}

function nodeTypes(editor: Editor) {
	return editor.state.doc.content.content.map(node => node.type.name)
}

beforeEach(() => {
	setActivePinia(createPinia())
})

afterEach(() => {
	document.body.innerHTML = ''
})

describe('Enter on a selection that starts at the beginning of a block', () => {
	it('does not throw when the selection ends in another paragraph', () => {
		const editor = createEditor('<p>abc</p><p>def</p>')
		// From the start of "abc" to the start of "def".
		select(editor, 1, 6)

		expect(() => pressEnter(editor)).not.toThrow()

		expect(nodeTypes(editor)).toEqual(['paragraph', 'paragraph'])
		expect(editor.getText()).toBe('\n\ndef')
	})

	it('does not throw when the selection ends inside a list', () => {
		const editor = createEditor('<p>abc</p><ul><li>def</li></ul>')
		// From the start of "abc" to the start of "def".
		select(editor, 1, 8)

		expect(() => pressEnter(editor)).not.toThrow()

		expect(editor.getText()).toContain('def')
	})

	it('leaves a selection inside a single block to the regular split', () => {
		const editor = createEditor('<p>abcdef</p>')
		select(editor, 2, 4)

		pressEnter(editor)

		expect(nodeTypes(editor)).toEqual(['paragraph', 'paragraph'])
		expect(editor.getText()).toBe('a\n\ndef')
	})
})
