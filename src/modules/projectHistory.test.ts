import {test, expect, vi} from 'vitest'
import {getHistory, removeProjectFromHistory, saveProjectToHistory} from './projectHistory'

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

test('store project in history', () => {
	let saved = {}
	Storage.prototype.getItem = vi.fn(() => null)
	Storage.prototype.setItem = vi.fn((key, projects) => {
		saved = projects
	})

	saveProjectToHistory({id: 1})
	expect(saved).toBe('[{"id":1}]')
})

test('store only the last 5 projects in history', () => {
	let saved: string | null = null
	Storage.prototype.getItem = vi.fn(() => saved)
	Storage.prototype.setItem = vi.fn((key: string, projects: string) => {
		saved = projects
	})

	saveProjectToHistory({id: 1})
	saveProjectToHistory({id: 2})
	saveProjectToHistory({id: 3})
	saveProjectToHistory({id: 4})
	saveProjectToHistory({id: 5})
	saveProjectToHistory({id: 6})
	expect(saved).toBe('[{"id":6},{"id":5},{"id":4},{"id":3},{"id":2}]')
})

test('don\'t store the same project twice', () => {
	let saved: string | null = null
	Storage.prototype.getItem = vi.fn(() => saved)
	Storage.prototype.setItem = vi.fn((key: string, projects: string) => {
		saved = projects
	})

	saveProjectToHistory({id: 1})
	saveProjectToHistory({id: 1})
	expect(saved).toBe('[{"id":1}]')
})

test('move a project to the beginning when storing it multiple times', () => {
	let saved: string | null = null
	Storage.prototype.getItem = vi.fn(() => saved)
	Storage.prototype.setItem = vi.fn((key: string, projects: string) => {
		saved = projects
	})

	saveProjectToHistory({id: 1})
	saveProjectToHistory({id: 2})
	saveProjectToHistory({id: 1})
	expect(saved).toBe('[{"id":1},{"id":2}]')
})

test('remove project from history', () => {
	let saved: string | null = '[{"id": 1}]'
	Storage.prototype.getItem = vi.fn(() => null)
	Storage.prototype.setItem = vi.fn((key: string, projects: string) => {
		saved = projects
	})
	Storage.prototype.removeItem = vi.fn((key: string) => {
		saved = null
	})

	removeProjectFromHistory({id: 1})
	expect(saved).toBeNull()
})
