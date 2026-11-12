import {describe, it, expect} from 'vitest'
import {htmlToPreviewText} from './htmlToPreviewText'

describe('htmlToPreviewText', () => {
	it('strips tags', () => {
		expect(htmlToPreviewText('<p>Hello <strong>world</strong></p>')).toBe('Hello world')
	})

	it('returns empty string for empty input', () => {
		expect(htmlToPreviewText('')).toBe('')
	})

	it('decodes entities', () => {
		expect(htmlToPreviewText('<p>a &amp; b</p>')).toBe('a & b')
	})

	it('handles many unclosed < quickly', () => {
		const start = performance.now()
		htmlToPreviewText('<'.repeat(200_000))
		expect(performance.now() - start).toBeLessThan(500)
	})
})
