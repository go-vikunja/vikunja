import {describe, expect, it} from 'vitest'
import {standaloneLinkUrls} from './standaloneLinkUrls'

const url = 'https://example.com/article'
const link = `<a href="${url}">${url}</a>`

describe('standaloneLinkUrls', () => {
	it('deduplicates bare URLs on their own line', () => {
		expect(standaloneLinkUrls(`<p>${link}</p><p>${link}</p>`, [])).toEqual([url])
	})
	it('skips inline, named and code links', () => {
		expect(standaloneLinkUrls(`<p>See ${link}</p><p><a href="${url}">Article</a></p>` +
			`<pre>${link}</pre><p><code>${link}</code></p>`, [])).toEqual([])
	})
	it('recognizes lines separated by hard breaks', () => {
		expect(standaloneLinkUrls(`<p>Before<br>${link}<br>After</p>`, [])).toEqual([url])
	})
	it('excludes both the current and configured application URLs', () => {
		const internal = 'https://tasks.example.com/vikunja/tasks/1'
		expect(standaloneLinkUrls(`<p><a href="${internal}">${internal}</a></p>`, [
			'https://tasks.example.com/vikunja/',
		])).toEqual([])
		expect(standaloneLinkUrls(`<p>${link}</p>`, ['https://example.com/app/'])).toEqual([url])
	})
	it('compares URL boundaries instead of matching a hostname prefix', () => {
		expect(standaloneLinkUrls(`<p>${link}</p>`, ['https://example.com.evil/'])).toEqual([url])
	})
	it('rejects credentials and non HTTP URLs', () => {
		for (const href of ['javascript:alert(1)', 'ftp://example.com', '/tasks/1', 'https://user:secret@example.com']) {
			expect(standaloneLinkUrls(`<p><a href="${href}">${href}</a></p>`, [])).toEqual([])
		}
	})
	it('preserves the linked section when opening a preview', () => {
		const href = `${url}#section`
		expect(standaloneLinkUrls(`<p><a href="${href}">${href}</a></p>`, [])).toEqual([href])
	})
})
