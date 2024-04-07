import {test, expect, vi} from 'vitest'
import {getHistory, removeProjectFromHistory, saveProjectToHistory} from './projectHistory'

test('return an empty history when none was saved', () => {
	vi.spyOn(localStorage, 'getItem').mockImplementation(() => null)
	const h = getHistory()
	expect(h).toStrictEqual([])
})

test('return a saved history', () => {
	const saved = [{id: 1}, {id: 2}]
	vi.spyOn(localStorage, 'getItem').mockImplementation(() => JSON.stringify(saved))

	const h = getHistory()
	expect(h).toStrictEqual(saved)
})

test('store project in history', () => {
	let saved = {}
	vi.spyOn(localStorage, 'getItem').mockImplementation(() => null)
	vi.spyOn(localStorage, 'setItem').mockImplementation((key: string, projects: string) => {
		saved = projects
	})

	saveProjectToHistory({id: 1})
	expect(saved).toBe('[{"id":1}]')
})

test('store only the last 6 projects in history', () => {
	let saved: string | null = null
	vi.spyOn(localStorage, 'getItem').mockImplementation(() => saved)
	vi.spyOn(localStorage, 'setItem').mockImplementation((key: string, projects: string) => {
		saved = projects
	})

	saveProjectToHistory({id: 1})
	saveProjectToHistory({id: 2})
	saveProjectToHistory({id: 3})
	saveProjectToHistory({id: 4})
	saveProjectToHistory({id: 5})
	saveProjectToHistory({id: 6})
	saveProjectToHistory({id: 7})
	expect(saved).toBe('[{"id":7},{"id":6},{"id":5},{"id":4},{"id":3},{"id":2}]')
})

test('don\'t store the same project twice', () => {
	let saved: string | null = null
	vi.spyOn(localStorage, 'getItem').mockImplementation(() => saved)
	vi.spyOn(localStorage, 'setItem').mockImplementation((key: string, projects: string) => {
		saved = projects
	})

	saveProjectToHistory({id: 1})
	saveProjectToHistory({id: 1})
	expect(saved).toBe('[{"id":1}]')
})

test('move a project to the beginning when storing it multiple times', () => {
	let saved: string | null = null
	vi.spyOn(localStorage, 'getItem').mockImplementation(() => saved)
	vi.spyOn(localStorage, 'setItem').mockImplementation((key: string, projects: string) => {
		saved = projects
	})

	saveProjectToHistory({id: 1})
	saveProjectToHistory({id: 2})
	saveProjectToHistory({id: 1})
	expect(saved).toBe('[{"id":1},{"id":2}]')
})

test('remove project from history', () => {
	let saved: string | null = '[{"id": 1}]'
	vi.spyOn(localStorage, 'getItem').mockImplementation(() => saved)
	vi.spyOn(localStorage, 'setItem').mockImplementation((key: string, projects: string) => {
		saved = projects
	})
	vi.spyOn(localStorage, 'removeItem').mockImplementation((key: string) => {
		saved = null
	})

	removeProjectFromHistory({id: 1})
	expect(saved).toBeNull()
})
