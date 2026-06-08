import {describe, it, expect} from 'vitest'
import {Editor} from '@tiptap/core'
import StarterKit from '@tiptap/starter-kit'
import {BlockquoteWithCommentId} from './blockquoteWithCommentId'

describe('BlockquoteWithCommentId extension', () => {
	const createEditor = (content: string = '') => {
		return new Editor({
			extensions: [
				StarterKit.configure({blockquote: false}),
				BlockquoteWithCommentId,
			],
			content,
		})
	}

	it('preserves data-comment-id through setContent → getHTML round-trip', () => {
		const editor = createEditor('<blockquote data-comment-id="42"><p>hi</p></blockquote>')

		const html = editor.getHTML()
		expect(html).toContain('data-comment-id="42"')

		editor.destroy()
	})

	it('renders a plain blockquote (no attribute) unchanged', () => {
		const editor = createEditor('<blockquote><p>just a quote</p></blockquote>')

		const html = editor.getHTML()
		expect(html).toContain('<blockquote>')
		expect(html).not.toContain('data-comment-id')

		editor.destroy()
	})

	it('preserves nested rich content inside the blockquote', () => {
		const editor = createEditor(
			'<blockquote data-comment-id="7"><p>this is <strong>bold</strong> text</p></blockquote>',
		)

		const html = editor.getHTML()
		expect(html).toContain('data-comment-id="7"')
		expect(html).toContain('<strong>bold</strong>')

		editor.destroy()
	})

	it('drops a malformed data-comment-id (non-integer)', () => {
		const editor = createEditor('<blockquote data-comment-id="abc"><p>x</p></blockquote>')

		const html = editor.getHTML()
		expect(html).not.toContain('data-comment-id')

		editor.destroy()
	})

	it('drops a non-positive data-comment-id', () => {
		const editor = createEditor('<blockquote data-comment-id="0"><p>x</p></blockquote>')

		const html = editor.getHTML()
		expect(html).not.toContain('data-comment-id')

		editor.destroy()
	})
})
