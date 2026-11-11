import {describe, expect, it, vi} from 'vitest'
import {fetchAllPages} from './fetchAllPages'

describe('fetchAllPages', () => {
	it('requests each page until total_pages and concatenates items', async () => {
		const fetchPage = vi.fn()
			.mockResolvedValueOnce({items: [1], total_pages: 3})
			.mockResolvedValueOnce({items: null, total_pages: 3})
			.mockResolvedValueOnce({items: [3], total_pages: 3})
		expect(await fetchAllPages(fetchPage)).toEqual([1, 3])
		expect(fetchPage.mock.calls).toEqual([[1], [2], [3]])
	})
})
