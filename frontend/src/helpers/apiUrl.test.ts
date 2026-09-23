import {describe, it, expect, afterEach} from 'vitest'
import {getApiBaseUrl, getLegacyApiBaseUrl, InvalidApiUrlProvidedError} from './apiUrl'

describe('API base URL', () => {
	const originalApiUrl = window.API_URL

	afterEach(() => {
		window.API_URL = originalApiUrl
	})

	it.each([
		['/api/v1', '/api/v2'],
		['/api/v2', '/api/v2'],
		['https://api.example/root/api/v2/', 'https://api.example/root/api/v2'],
		['https://api.example/root/api/v1', 'https://api.example/root/api/v2'],
		['https://api.example/api/v2/tenant/api/v1/', 'https://api.example/api/v2/tenant/api/v2'],
		['https://api.example/custom', 'https://api.example/custom'],
		['https://api.example/custom/', 'https://api.example/custom'],
	])('normalizes %s without changing the deployment prefix', (input, expected) => {
		window.API_URL = input
		expect(getApiBaseUrl()).toBe(expected)
	})

	it.each([
		['', 'empty'],
		[undefined as unknown as string, 'undefined'],
		['/', 'the bare origin'],
	])('rejects %o (%s) instead of falling back to the frontend origin', input => {
		window.API_URL = input
		expect(() => getApiBaseUrl()).toThrow(InvalidApiUrlProvidedError)
		expect(() => getLegacyApiBaseUrl()).toThrow(InvalidApiUrlProvidedError)
	})

	it.each([
		['/api/v2', '/api/v1'],
		['/api/v1', '/api/v1'],
		['https://api.example/custom', 'https://api.example/custom'],
	])('downgrades %s back to the legacy base', (input, expected) => {
		window.API_URL = input
		expect(getLegacyApiBaseUrl()).toBe(expected)
	})
})
