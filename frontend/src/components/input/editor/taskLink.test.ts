import {describe, it, expect, beforeEach} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {ref} from 'vue'
import {Editor} from '@tiptap/core'
import {Fragment, Slice} from '@tiptap/pm/model'
import {createEditorExtensions, type EditorExtensionDeps} from './editorExtensions'

const TASK_URL = 'http://localhost:3000/tasks/123'

// Inert deps so the editor runs TipTap.vue's real extension set, whose markdown
// paste handler and Link mark both compete with TaskLink for pastes.
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

// pasteText() leaves clipboardData empty, which the markdown paste handler bails
// on — go through the plugin chain the way ProseMirror does instead.
function paste(editor: Editor, text: string) {
	const event = {clipboardData: {getData: () => text, items: []}} as unknown as ClipboardEvent
	const slice = new Slice(Fragment.from(editor.schema.text(text)), 0, 0)
	editor.view.someProp('handlePaste', handler => handler(editor.view, event, slice))
}

beforeEach(() => {
	setActivePinia(createPinia())
})

describe('TaskLink extension', () => {
	const createEditor = (content: string = '') => {
		return new Editor({
			extensions: createEditorExtensions(stubDeps),
			content,
		})
	}

	const findTaskLinks = (editor: Editor) => {
		const found: {href: string}[] = []
		editor.state.doc.descendants(node => {
			if (node.type.name === 'taskLink') {
				found.push({href: node.attrs.href})
			}
		})
		return found
	}

	it('upgrades a same-origin task url anchor whose text equals its href', () => {
		const editor = createEditor(`<p><a href="${TASK_URL}">${TASK_URL}</a></p>`)

		expect(findTaskLinks(editor)).toEqual([{href: TASK_URL}])

		editor.destroy()
	})

	it('serializes back to the same plain anchor', () => {
		const html = `<p><a href="${TASK_URL}">${TASK_URL}</a></p>`
		const editor = createEditor(html)

		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})

	it('preserves target and rel attributes through the round-trip', () => {
		const html = `<p><a target="_blank" rel="noopener noreferrer nofollow" href="${TASK_URL}">${TASK_URL}</a></p>`
		const editor = createEditor(html)

		expect(findTaskLinks(editor)).toHaveLength(1)
		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})

	it('leaves a task url anchor with custom text as a plain link', () => {
		const html = `<p><a target="_blank" rel="noopener noreferrer nofollow" href="${TASK_URL}">see this task</a></p>`
		const editor = createEditor(html)

		expect(findTaskLinks(editor)).toHaveLength(0)
		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})

	it('leaves a non-task anchor as a plain link', () => {
		const html = '<p><a target="_blank" rel="noopener noreferrer nofollow" href="https://example.com/">https://example.com/</a></p>'
		const editor = createEditor(html)

		expect(findTaskLinks(editor)).toHaveLength(0)
		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})

	it('leaves an other-origin task url as a plain link', () => {
		const html = '<p><a target="_blank" rel="noopener noreferrer nofollow" href="https://other.example.com/tasks/123">https://other.example.com/tasks/123</a></p>'
		const editor = createEditor(html)

		expect(findTaskLinks(editor)).toHaveLength(0)
		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})

	it('leaves a task url anchor containing child markup as a plain link', () => {
		const html = `<p><a target="_blank" rel="noopener noreferrer nofollow" href="${TASK_URL}"><u>${TASK_URL}</u></a></p>`
		const editor = createEditor(html)

		expect(findTaskLinks(editor)).toHaveLength(0)
		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})

	it('leaves a task url anchor with extra attributes as a plain link', () => {
		const html = `<p><a target="_blank" rel="noopener noreferrer nofollow" href="${TASK_URL}" title="x">${TASK_URL}</a></p>`
		const editor = createEditor(html)

		expect(findTaskLinks(editor)).toHaveLength(0)
		expect(editor.getHTML()).toBe(html)

		editor.destroy()
	})

	it('creates one node per link, even for the same task', () => {
		const editor = createEditor(
			`<p><a href="${TASK_URL}">${TASK_URL}</a> and <a href="${TASK_URL}">${TASK_URL}</a></p>`,
		)

		expect(findTaskLinks(editor)).toHaveLength(2)

		editor.destroy()
	})
})

describe('TaskLink paste handling', () => {
	const createEditor = (content: string = '<p></p>') => {
		return new Editor({
			extensions: createEditorExtensions(stubDeps),
			content,
		})
	}

	const countTaskLinks = (editor: Editor) => {
		let count = 0
		editor.state.doc.descendants(node => {
			if (node.type.name === 'taskLink') {
				count++
			}
		})
		return count
	}

	it('inserts a taskLink node when pasting a bare task url', () => {
		const editor = createEditor()
		editor.commands.focus('end')

		editor.view.pasteText(TASK_URL)

		expect(countTaskLinks(editor)).toBe(1)
		expect(editor.getHTML()).toBe(`<p><a target="_blank" rel="noopener noreferrer nofollow" href="${TASK_URL}">${TASK_URL}</a></p>`)

		editor.destroy()
	})

	it('does not pill-ify a task url pasted as part of a longer text', () => {
		const editor = createEditor()
		editor.commands.focus('end')

		editor.view.pasteText(`see ${TASK_URL} please`)

		expect(countTaskLinks(editor)).toBe(0)
		expect(editor.getText()).toBe(`see ${TASK_URL} please`)

		editor.destroy()
	})

	it('leaves pastes inside code blocks alone', () => {
		const editor = createEditor('<pre><code></code></pre>')
		editor.commands.focus('end')

		editor.view.pasteText(TASK_URL)

		expect(countTaskLinks(editor)).toBe(0)
		expect(editor.getHTML()).toBe(`<pre><code>${TASK_URL}</code></pre><p></p>`)

		editor.destroy()
	})

	it('does not pill-ify when text is selected (link mark applies instead)', () => {
		const editor = createEditor('<p>hello</p>')
		editor.commands.setTextSelection({from: 1, to: 6})

		editor.view.pasteText(TASK_URL)

		expect(countTaskLinks(editor)).toBe(0)
		expect(editor.getHTML()).toBe(`<p><a target="_blank" rel="noopener noreferrer nofollow" href="${TASK_URL}">hello</a></p>`)
		expect(editor.getText()).toBe('hello')

		editor.destroy()
	})

	it('pill-ifies a task url containing markdown syntax characters', () => {
		const editor = createEditor()
		editor.commands.focus('end')
		const url = `${TASK_URL}#comment-1`

		paste(editor, url)

		expect(countTaskLinks(editor)).toBe(1)
		expect(editor.getHTML()).toBe(`<p><a target="_blank" rel="noopener noreferrer nofollow" href="${url}">${url}</a></p>`)

		editor.destroy()
	})

	it('ignores non-task urls', () => {
		const editor = createEditor()
		editor.commands.focus('end')

		editor.view.pasteText('https://example.com/')

		expect(countTaskLinks(editor)).toBe(0)

		editor.destroy()
	})
})
