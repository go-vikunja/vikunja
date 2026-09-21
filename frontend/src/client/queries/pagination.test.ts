import {describe, expect, it} from 'vitest'
import {API_MAX_PER_PAGE, pageSizeFor, totalPagesFor} from './pagination'

describe('totalPagesFor', () => {
	it('rounds a partial page up', () => {
		expect(totalPagesFor({per_page: 50, total_pages: 1}, 101)).toBe(3)
	})

	it('keeps the cached total_pages when per_page is 0', () => {
		expect(totalPagesFor({per_page: 0, total_pages: 7}, 101)).toBe(7)
	})

	it('does not add a page for an exact multiple', () => {
		expect(totalPagesFor({per_page: 50, total_pages: 1}, 100)).toBe(2)
	})
})

describe('pageSizeFor', () => {
	it('clamps 0 up to 1', () => {
		expect(pageSizeFor(0)).toBe(1)
	})

	it('keeps a configured size within the bound', () => {
		expect(pageSizeFor(50)).toBe(50)
	})

	it('clamps above the api maximum', () => {
		expect(pageSizeFor(5000)).toBe(API_MAX_PER_PAGE)
	})
})
