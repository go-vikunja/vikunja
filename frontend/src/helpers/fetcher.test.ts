import {beforeEach, describe, expect, it} from 'vitest'

import {apiV2Url, getApiV2BaseUrl} from './fetcher'

describe('v2 API URLs', () => {
	beforeEach(() => {
		window.API_URL = 'https://api.example.com/root/api/v1'
	})

	it('derives the v2 base URL from the configured API URL', () => {
		expect(getApiV2BaseUrl()).toBe('https://api.example.com/root/api/v2/')
	})

	it('builds v2 endpoint URLs from the shared base URL', () => {
		expect(apiV2Url('labels?sort_by=title')).toBe('https://api.example.com/root/api/v2/labels?sort_by=title')
	})
})
