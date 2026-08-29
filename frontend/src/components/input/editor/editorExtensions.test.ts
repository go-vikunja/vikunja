import {describe, it, expect, afterEach, beforeEach, vi} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {nextTick, ref} from 'vue'
import {Editor} from '@tiptap/core'
import {createEditorExtensions, type EditorExtensionDeps} from './editorExtensions'

const API_URL = 'http://localhost:3456/api/v1'
window.API_URL = API_URL

const ATTACHMENT_URL = `${API_URL}/tasks/5/attachments/9`

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
		attachmentService: {getBlobUrl: vi.fn(async () => 'blob:real-attachment')} as never,
	}

	editor = new Editor({
		element: holder,
		extensions: createEditorExtensions(deps),
		content,
	})

	return {editor, holder}
}

async function settle() {
	await nextTick()
	await new Promise(resolve => setTimeout(resolve, 0))
	await nextTick()
}

beforeEach(() => {
	setActivePinia(createPinia())
})

afterEach(() => {
	document.body.innerHTML = ''
})

describe('CustomImage attachment id', () => {
	it('resolves the blob url for a freshly inserted image', async () => {
		const {editor} = createEditor('<p>hi</p>')

		editor.chain().setImage({src: ATTACHMENT_URL}).run()
		await settle()

		const img = editor.view.dom.querySelector('img')!
		expect(img.getAttribute('data-src')).toBe(ATTACHMENT_URL)
		expect(img.id).toBe('tiptap-image-5-9')
		expect(img.src).toBe('blob:real-attachment')
	})

	it('resolves the blob url for a reloaded stored description', async () => {
		const stored = `<p><img src="#" data-src="${ATTACHMENT_URL}" id="tiptap-image-5-9"></p>`
		const {editor} = createEditor(stored)
		await settle()

		const img = editor.view.dom.querySelector('img')!
		expect(img.id).toBe('tiptap-image-5-9')
		expect(img.src).toBe('blob:real-attachment')
	})

	it('resolves the blob url for a legacy stored description without data-src', async () => {
		const {editor} = createEditor(`<p><img src="${ATTACHMENT_URL}"></p>`)
		await settle()

		const img = editor.view.dom.querySelector('img')!
		expect(img.id).toBe('tiptap-image-5-9')
		expect(img.src).toBe('blob:real-attachment')
	})

	it('does not let a planted id hijack another image\'s blob url', async () => {
		const stored = `<p><img src="https://attacker.example/x.png" id="tiptap-image-5-9">` +
			`<img src="#" data-src="${ATTACHMENT_URL}" id="tiptap-image-5-9"></p>`
		const {editor} = createEditor(stored)
		await settle()

		const [planted, real] = Array.from(editor.view.dom.querySelectorAll('img'))
		expect(planted.id).toBe('')
		expect(planted.src).toBe('https://attacker.example/x.png')
		expect(real.src).toBe('blob:real-attachment')
	})

	it('resolves the blob url independently for two editors with the same attachment', async () => {
		const stored = `<p><img src="#" data-src="${ATTACHMENT_URL}" id="tiptap-image-5-9"></p>`
		const first = createEditor(stored)
		const second = createEditor(stored)
		await settle()

		expect(first.editor.view.dom.querySelector('img')!.src).toBe('blob:real-attachment')
		expect(second.editor.view.dom.querySelector('img')!.src).toBe('blob:real-attachment')
	})

	it('round trips the rendered html', async () => {
		const {editor} = createEditor(`<p><img src="${ATTACHMENT_URL}"></p>`)
		await settle()

		const html = editor.getHTML()
		expect(html).toContain(`data-src="${ATTACHMENT_URL}"`)
		expect(html).toContain('id="tiptap-image-5-9"')

		const reloaded = createEditor(html)
		await settle()
		expect(reloaded.editor.view.dom.querySelector('img')!.src).toBe('blob:real-attachment')
	})
})
