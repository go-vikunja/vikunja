import {describe, it, expect} from 'vitest'

import {htmlHasFormatting} from './clipboardHtml'

describe('htmlHasFormatting', () => {
	it.each([
		['an empty string', ''],
		['bare text', 'hello - world'],
		['syntax highlighted span soup', '<div style="color:#d4d4d4"><span># h</span></div>'],
		['plain paragraphs', '<p>a</p><p>b</p>'],
	])('reports no formatting for %s', (_name, html) => {
		expect(htmlHasFormatting(html)).toBe(false)
	})

	it.each([
		['a link', '<span>see <a href="https://vikunja.io">docs</a></span>'],
		['bold text', 'a <strong>b</strong>'],
		['a heading', '<h2>title</h2>'],
		['a list', '<ul><li>a</li></ul>'],
		['a table', '<table><tr><td>a</td></tr></table>'],
		['an image', '<img src="https://vikunja.io/logo.png">'],
		['code', '<code>a - b</code>'],
	])('reports formatting for %s', (_name, html) => {
		expect(htmlHasFormatting(html)).toBe(true)
	})
})
