import {describe, expect, it} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	normalizePageNumber,
	removeFromPages,
	toPaginated,
	totalPagesFor,
} from './pagination'

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

describe('toPaginated', () => {
	it('fills missing envelope fields and falls back to the requested page', () => {
		expect(toPaginated({items: null}, 3)).toEqual({
			items: [],
			page: 3,
			per_page: 0,
			total: 0,
			total_pages: 0,
		})
	})
})

describe('normalizePageNumber', () => {
	it.each([
		['2', 2],
		['abc', 1],
		['0', 1],
		['-3', 1],
		['1.5', 1],
		[undefined, 1],
		[7, 7],
		[null, 1],
		['', 1],
		[['1', '2'], 1],
	])('normalizes page %s to %i', (raw, expected) => {
		expect(normalizePageNumber(raw)).toBe(expected)
	})
})

describe('removeFromPages', () => {
	type Item = {id: number}
	const page = (items: Item[], page: number, total: number) => ({
		items,
		page,
		per_page: 2,
		total,
		total_pages: Math.ceil(total / 2),
	})

	it('drops the item and decrements every page of the scope that held it', () => {
		const client = new QueryClient()
		client.setQueryData(['things', 'list', 'a', 1], page([{id: 1}, {id: 2}], 1, 3))
		client.setQueryData(['things', 'list', 'a', 2], page([{id: 3}], 2, 3))
		client.setQueryData(['things', 'list', 'b', 1], page([{id: 4}, {id: 5}], 1, 3))

		removeFromPages<Item>(client, ['things', 'list'], item => item.id === 3)

		expect(client.getQueryData(['things', 'list', 'a', 1])).toEqual(page([{id: 1}, {id: 2}], 1, 2))
		expect(client.getQueryData(['things', 'list', 'a', 2])).toEqual(page([], 2, 2))
		expect(client.getQueryData(['things', 'list', 'b', 1])).toEqual(page([{id: 4}, {id: 5}], 1, 3))
	})

	it('leaves the cache untouched when no cached page holds the item', () => {
		const client = new QueryClient()
		client.setQueryData(['things', 'list', 'a', 1], page([{id: 1}], 1, 3))

		removeFromPages<Item>(client, ['things', 'list'], item => item.id === 9)

		expect(client.getQueryData(['things', 'list', 'a', 1])).toEqual(page([{id: 1}], 1, 3))
	})
})
