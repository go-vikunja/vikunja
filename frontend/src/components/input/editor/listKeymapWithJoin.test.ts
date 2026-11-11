import {describe, it, expect, afterEach} from 'vitest'
import {Editor} from '@tiptap/core'
import StarterKit from '@tiptap/starter-kit'
import {TaskList} from '@tiptap/extension-list'
import {TaskItemWithId} from './taskItemWithId'
import {ListKeymapWithJoin} from './listKeymapWithJoin'

let editor: Editor | null = null

function createEditor(content: string): Editor {
	editor = new Editor({
		extensions: [
			StarterKit.configure({listKeymap: false}),
			ListKeymapWithJoin,
			TaskList,
			TaskItemWithId.configure({nested: true}),
		],
		content,
	})

	return editor
}

function setCursorBefore(editor: Editor, text: string) {
	let pos = -1
	editor.state.doc.descendants((node, nodePos) => {
		if (pos === -1 && node.isText && node.text === text) {
			pos = nodePos
		}
	})
	expect(pos).toBeGreaterThan(-1)
	editor.commands.setTextSelection(pos)
}

function pressBackspace(editor: Editor) {
	const event = new KeyboardEvent('keydown', {key: 'Backspace', bubbles: true, cancelable: true})
	editor.view.someProp('handleKeyDown', handler => handler(editor.view, event))
}

// StarterKit's TrailingNode always keeps an empty paragraph at the end of the document.
function html(editor: Editor): string {
	return editor.getHTML()
		.replace(/ data-task-id="[^"]*"/g, '')
		.replace(/<p><\/p>$/, '')
}

afterEach(() => {
	editor?.destroy()
	editor = null
})

describe('ListKeymapWithJoin', () => {
	it('joins a bullet list item into the one above instead of splitting the list', () => {
		const editor = createEditor('<ul><li><p>one</p></li><li><p>two</p></li><li><p>three</p></li></ul>')

		setCursorBefore(editor, 'two')
		pressBackspace(editor)

		expect(html(editor)).toBe('<ul><li><p>onetwo</p></li><li><p>three</p></li></ul>')
	})

	it('joins an ordered list item into the one above', () => {
		const editor = createEditor('<ol><li><p>one</p></li><li><p>two</p></li></ol>')

		setCursorBefore(editor, 'two')
		pressBackspace(editor)

		expect(html(editor)).toBe('<ol><li><p>onetwo</p></li></ol>')
	})

	it('joins a task list item into the one above', () => {
		const editor = createEditor('<ul data-type="taskList"><li data-type="taskItem" data-checked="false"><p>one</p></li><li data-type="taskItem" data-checked="false"><p>two</p></li></ul>')

		setCursorBefore(editor, 'two')
		pressBackspace(editor)

		expect(html(editor)).toContain('<p>onetwo</p>')
		expect(html(editor)).not.toContain('<p>two</p>')
	})

	it('joins into the last item of a preceding nested list', () => {
		const editor = createEditor('<ul><li><p>one</p><ul><li><p>sub</p></li></ul></li><li><p>two</p></li></ul>')

		setCursorBefore(editor, 'two')
		pressBackspace(editor)

		expect(html(editor)).toBe('<ul><li><p>one</p><ul><li><p>subtwo</p></li></ul></li></ul>')
	})

	it('still lifts the first item out of the list', () => {
		const editor = createEditor('<ul><li><p>one</p></li><li><p>two</p></li></ul>')

		setCursorBefore(editor, 'one')
		pressBackspace(editor)

		expect(html(editor)).toBe('<p>one</p><ul><li><p>two</p></li></ul>')
	})

	it('still outdents the first item of a nested list', () => {
		const editor = createEditor('<ul><li><p>one</p><ul><li><p>sub</p></li></ul></li></ul>')

		setCursorBefore(editor, 'sub')
		pressBackspace(editor)

		expect(html(editor)).toBe('<ul><li><p>one</p></li><li><p>sub</p></li></ul>')
	})

	it('does not join when the cursor is not at the start of the item', () => {
		const editor = createEditor('<ul><li><p>one</p></li><li><p>two</p></li></ul>')

		setCursorBefore(editor, 'two')
		editor.commands.setTextSelection(editor.state.selection.from + 1)
		pressBackspace(editor)

		// Deleting the character itself is the browser's job, so the document must stay untouched.
		expect(html(editor)).toBe('<ul><li><p>one</p></li><li><p>two</p></li></ul>')
	})

	it('does not join from a second paragraph inside a list item', () => {
		const editor = createEditor('<ul><li><p>one</p></li><li><p>two</p><p>three</p></li></ul>')

		setCursorBefore(editor, 'three')
		pressBackspace(editor)

		expect(html(editor)).toBe('<ul><li><p>one</p></li><li><p>twothree</p></li></ul>')
	})
})
