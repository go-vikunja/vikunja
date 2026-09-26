import {beforeEach, describe, expect, it, vi} from 'vitest'
import {CancelledError} from '@tanstack/vue-query'

import {queryClient} from './queryClient'
import {error} from '@/message'

vi.mock('@/message', () => ({error: vi.fn()}))

function failingQuery(queryKey: string[], cause: unknown) {
	return queryClient.fetchQuery({
		queryKey,
		queryFn: async () => {
			throw cause
		},
		retry: false,
	})
}

describe('queryClient', () => {
	beforeEach(() => {
		queryClient.clear()
		vi.mocked(error).mockClear()
	})

	it('reports a failed read', async () => {
		const cause = {
			status: 500,
			detail: 'boom',
		}

		await expect(failingQuery(['fails'], cause)).rejects.toBe(cause)

		expect(error).toHaveBeenCalledExactlyOnceWith(cause)
	})

	it('stays silent for a cancelled read', async () => {
		await expect(failingQuery(['cancelled'], new CancelledError())).rejects.toThrow()

		expect(error).not.toHaveBeenCalled()
	})

	it('stays silent for a request context abort', async () => {
		const cause = new DOMException('Client request context changed', 'AbortError')

		await expect(failingQuery(['aborted'], cause)).rejects.toBe(cause)

		expect(error).not.toHaveBeenCalled()
	})

	it('stays silent for a query that reports its own failures', async () => {
		const cause = {
			status: 404,
			detail: 'gone',
		}

		await expect(queryClient.fetchQuery({
			queryKey: ['handled'],
			queryFn: async () => {
				throw cause
			},
			retry: false,
			meta: {handlesError: true},
		})).rejects.toBe(cause)

		expect(error).not.toHaveBeenCalled()
	})
})
