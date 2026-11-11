import {describe, it, expect} from 'vitest'

import {parseQuickAddPrefill, quickAddPrefillDescription} from './quickAddPrefill'

describe('parseQuickAddPrefill', () => {
	it('reads title, url and description', () => {
		expect(parseQuickAddPrefill({title: ' Page ', url: 'https://example.com', description: 'notes'}))
			.toEqual({title: 'Page', url: 'https://example.com', description: 'notes'})
	})

	it('returns null without any prefill param', () => {
		expect(parseQuickAddPrefill({labels: '1'})).toBeNull()
	})

	it('defaults missing params to empty strings', () => {
		expect(parseQuickAddPrefill({title: 'Page'})).toEqual({title: 'Page', url: '', description: ''})
	})

	it('uses the first value of a repeated param', () => {
		expect(parseQuickAddPrefill({title: ['first', 'second']})?.title).toBe('first')
	})

	it('moves a url shared as description into url', () => {
		expect(parseQuickAddPrefill({title: 'Page', description: 'https://example.com/a'}))
			.toEqual({title: 'Page', url: 'https://example.com/a', description: ''})
	})

	it('keeps a non-url description', () => {
		expect(parseQuickAddPrefill({description: 'not a url'})).toEqual({title: '', url: '', description: 'not a url'})
	})

	it('drops a description that repeats the url', () => {
		expect(parseQuickAddPrefill({url: 'https://example.com', description: 'https://example.com'})?.description).toBe('')
	})
})

describe('quickAddPrefillDescription', () => {
	it('is empty without url and description', () => {
		expect(quickAddPrefillDescription({url: '', description: ''})).toBe('')
	})

	it('links an http url', () => {
		expect(quickAddPrefillDescription({url: 'https://example.com/?a=1&b=2', description: ''}))
			.toBe('<p><a href="https://example.com/?a=1&amp;b=2">https://example.com/?a=1&amp;b=2</a></p>')
	})

	it('does not link a non-http url', () => {
		expect(quickAddPrefillDescription({url: 'javascript:alert(1)', description: ''}))
			.toBe('<p>javascript:alert(1)</p>')
	})

	it('keeps markup-like text as text and puts each line in a paragraph', () => {
		expect(quickAddPrefillDescription({url: '', description: '<img src=x onerror=alert(1)>\nsecond'}))
			.toBe('<p>&lt;img src=x onerror=alert(1)&gt;</p><p>second</p>')
	})

	it('puts the url before the description', () => {
		expect(quickAddPrefillDescription({url: 'https://example.com', description: 'notes'}))
			.toBe('<p><a href="https://example.com">https://example.com</a></p><p>notes</p>')
	})
})
