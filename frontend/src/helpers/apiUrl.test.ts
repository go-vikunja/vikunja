import {describe, it, expect} from 'vitest'
import {getApiBaseUrl} from './apiUrl'

describe('API base URL', () => {
	it.each([
		['/api/v2', '/api/v2/'],
		['https://api.example/root/api/v2/', 'https://api.example/root/api/v2/'],
		['https://api.example/root/api/v1', 'https://api.example/root/api/v2/'],
		['https://api.example/api/v2/tenant/api/v1/', 'https://api.example/api/v2/tenant/api/v2/'],
		['https://api.example/custom', 'https://api.example/custom/'],
	])('normalizes %s without changing the deployment prefix', (input, expected) => {
		window.API_URL = input
		expect(getApiBaseUrl()).toBe(expected)
	})
})
