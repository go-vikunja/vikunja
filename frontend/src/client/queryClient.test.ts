import {beforeEach, describe, expect, it, vi} from 'vitest'
import {CancelledError} from '@tanstack/vue-query'

import {queryClient} from './queryClient'
import {ApiError} from '@/client/problem'
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

	it('lets a query handle only the failures its predicate matches', async () => {
		const handled = {
			status: 404,
			detail: 'gone',
		}
		const unhandled = {
			status: 500,
			detail: 'boom',
		}
		const meta = {handlesError: (cause: unknown) => cause === handled}

		for (const cause of [handled, unhandled]) {
			await expect(queryClient.fetchQuery({
				queryKey: ['predicate', cause.status],
				queryFn: async () => {
					throw cause
				},
				retry: false,
				meta,
			})).rejects.toBe(cause)
		}

		expect(error).toHaveBeenCalledExactlyOnceWith(unhandled)
	})

	it('does not retry a read of a trashed project', async () => {
		const cause = new ApiError({
			status: 404,
			code: 3018,
			detail: 'in trash',
			details: {id: 1, deleted_at: '2026-01-01T00:00:00Z', can_restore: true},
		})
		const queryFn = vi.fn(async () => {
			throw cause
		})

		await expect(queryClient.fetchQuery({queryKey: ['trashed'], queryFn})).rejects.toBe(cause)

		expect(queryFn).toHaveBeenCalledOnce()
	})

	it('retries an ordinary failed read once', async () => {
		const cause = {
			status: 500,
			detail: 'boom',
		}
		const queryFn = vi.fn(async () => {
			throw cause
		})

		await expect(queryClient.fetchQuery({
			queryKey: ['flaky'],
			queryFn,
			retryDelay: 0,
		})).rejects.toBe(cause)

		expect(queryFn).toHaveBeenCalledTimes(2)
	})
})
