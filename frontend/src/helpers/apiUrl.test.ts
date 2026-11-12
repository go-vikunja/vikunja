import {describe, it, expect, afterEach} from 'vitest'
import {
	getApiBaseUrl,
	getApiRootUrl,
	InvalidApiUrlProvidedError,
	normalizeApiUrl,
} from './apiUrl'

describe('API base URL', () => {
	const originalApiUrl = window.API_URL

	afterEach(() => {
		window.API_URL = originalApiUrl
	})

	it.each([
		['', '', '/api/v2'],
		['/', '', '/api/v2'],
		['https://x', 'https://x', 'https://x/api/v2'],
		['https://x/', 'https://x', 'https://x/api/v2'],
		['https://x/prefix', 'https://x/prefix', 'https://x/prefix/api/v2'],
		['https://x/prefix/', 'https://x/prefix', 'https://x/prefix/api/v2'],
		['https://x/api/v1', 'https://x', 'https://x/api/v2'],
		['https://x/api/v2', 'https://x', 'https://x/api/v2'],
		['https://x/prefix/api/v1/', 'https://x/prefix', 'https://x/prefix/api/v2'],
		['/prefix/api/v2', '/prefix', '/prefix/api/v2'],
	])('derives the API URLs from the stored %o', (stored, root, base) => {
		window.API_URL = stored
		expect(normalizeApiUrl(stored)).toBe(root)
		expect(getApiRootUrl()).toBe(root)
		expect(getApiBaseUrl()).toBe(base)
	})

	it('rejects a missing API URL instead of guessing one', () => {
		window.API_URL = undefined as unknown as string
		expect(() => getApiBaseUrl()).toThrow(InvalidApiUrlProvidedError)
	})
})
