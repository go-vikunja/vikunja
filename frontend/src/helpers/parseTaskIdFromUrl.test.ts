import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'
import {useConfigStore} from '@/stores/config'
import {parseTaskIdFromUrl} from './parseTaskIdFromUrl'

describe('parseTaskIdFromUrl', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	afterEach(() => {
		vi.unstubAllEnvs()
	})

	it('parses an absolute same-origin task url', () => {
		expect(parseTaskIdFromUrl('http://localhost:3000/tasks/123')).toBe(123)
	})

	it('parses a relative task url', () => {
		expect(parseTaskIdFromUrl('/tasks/123')).toBe(123)
	})

	it('ignores query string and hash after the id', () => {
		expect(parseTaskIdFromUrl('http://localhost:3000/tasks/42?foo=bar')).toBe(42)
		expect(parseTaskIdFromUrl('http://localhost:3000/tasks/42#comment-1')).toBe(42)
	})

	it('rejects other origins', () => {
		expect(parseTaskIdFromUrl('https://other.example.com/tasks/123')).toBeNull()
		expect(parseTaskIdFromUrl('http://localhost:4000/tasks/123')).toBeNull()
	})

	it('rejects non-task routes and malformed ids', () => {
		expect(parseTaskIdFromUrl('/tasks/abc')).toBeNull()
		expect(parseTaskIdFromUrl('/tasks/12abc')).toBeNull()
		expect(parseTaskIdFromUrl('/tasks/12/edit')).toBeNull()
		expect(parseTaskIdFromUrl('/tasks/')).toBeNull()
		expect(parseTaskIdFromUrl('/tasks/0')).toBeNull()
		expect(parseTaskIdFromUrl('/projects/5')).toBeNull()
		expect(parseTaskIdFromUrl('/tasks/by/upcoming')).toBeNull()
		expect(parseTaskIdFromUrl('not a url')).toBeNull()
		expect(parseTaskIdFromUrl('')).toBeNull()
	})

	it('respects a non-root base path', () => {
		vi.stubEnv('BASE_URL', '/vikunja/')
		expect(parseTaskIdFromUrl('http://localhost:3000/vikunja/tasks/7')).toBe(7)
		expect(parseTaskIdFromUrl('/vikunja/tasks/7')).toBe(7)
		expect(parseTaskIdFromUrl('http://localhost:3000/tasks/7')).toBeNull()
	})

	it('accepts urls from the configured frontend url', () => {
		useConfigStore().frontendUrl = 'https://vikunja.example.com/'

		expect(parseTaskIdFromUrl('https://vikunja.example.com/tasks/9')).toBe(9)
		expect(parseTaskIdFromUrl('http://localhost:3000/tasks/9')).toBe(9)
		expect(parseTaskIdFromUrl('https://other.example.com/tasks/9')).toBeNull()
		expect(parseTaskIdFromUrl('https://vikunja.example.com.evil.com/tasks/9')).toBeNull()
		expect(parseTaskIdFromUrl('https://evil.vikunja.example.com/tasks/9')).toBeNull()
	})

	it('respects a path in the configured frontend url', () => {
		useConfigStore().frontendUrl = 'https://example.com/vikunja/'

		expect(parseTaskIdFromUrl('https://example.com/vikunja/tasks/9')).toBe(9)
		expect(parseTaskIdFromUrl('https://example.com/tasks/9')).toBeNull()
	})

	it('only accepts the current origin when no frontend url is configured', () => {
		expect(parseTaskIdFromUrl('http://localhost:3000/tasks/9')).toBe(9)
		expect(parseTaskIdFromUrl('https://vikunja.example.com/tasks/9')).toBeNull()
	})

	it('falls back to the current origin when the configured frontend url is malformed', () => {
		useConfigStore().frontendUrl = 'not a url'

		expect(parseTaskIdFromUrl('http://localhost:3000/tasks/9')).toBe(9)
		expect(parseTaskIdFromUrl('https://vikunja.example.com/tasks/9')).toBeNull()
	})
})
