import {test, expect, vi} from 'vitest'
import {getHistory, removeListFromHistory, saveListToHistory} from './listHistory'

test('return an empty history when none was saved', () => {
	Storage.prototype.getItem = vi.fn(() => null)
	const h = getHistory()
	expect(h).toStrictEqual([])
})

test('return a saved history', () => {
	const saved = [{id: 1}, {id: 2}]
	Storage.prototype.getItem = vi.fn(() => JSON.stringify(saved))

	const h = getHistory()
	expect(h).toStrictEqual(saved)
})

test('store list in history', () => {
	let saved = {}
	Storage.prototype.getItem = vi.fn(() => null)
	Storage.prototype.setItem = vi.fn((key, lists) => {
		saved = lists
	})

	saveListToHistory({id: 1})
	expect(saved).toBe('[{"id":1}]')
})

test('store only the last 5 lists in history', () => {
	let saved: string | null = null
	Storage.prototype.getItem = vi.fn(() => saved)
	Storage.prototype.setItem = vi.fn((key: string, lists: string) => {
		saved = lists
	})

	saveListToHistory({id: 1})
	saveListToHistory({id: 2})
	saveListToHistory({id: 3})
	saveListToHistory({id: 4})
	saveListToHistory({id: 5})
	saveListToHistory({id: 6})
	expect(saved).toBe('[{"id":6},{"id":5},{"id":4},{"id":3},{"id":2}]')
})

test('don\'t store the same list twice', () => {
	let saved: string | null = null
	Storage.prototype.getItem = vi.fn(() => saved)
	Storage.prototype.setItem = vi.fn((key: string, lists: string) => {
		saved = lists
	})

	saveListToHistory({id: 1})
	saveListToHistory({id: 1})
	expect(saved).toBe('[{"id":1}]')
})

test('move a list to the beginning when storing it multiple times', () => {
	let saved: string | null = null
	Storage.prototype.getItem = vi.fn(() => saved)
	Storage.prototype.setItem = vi.fn((key: string, lists: string) => {
		saved = lists
	})

	saveListToHistory({id: 1})
	saveListToHistory({id: 2})
	saveListToHistory({id: 1})
	expect(saved).toBe('[{"id":1},{"id":2}]')
})

test('remove list from history', () => {
	let saved: string | null = '[{"id": 1}]'
	Storage.prototype.getItem = vi.fn(() => null)
	Storage.prototype.setItem = vi.fn((key: string, lists: string) => {
		saved = lists
	})
	Storage.prototype.removeItem = vi.fn((key: string) => {
		saved = null
	})

	removeListFromHistory({id: 1})
	expect(saved).toBeNull()
})
