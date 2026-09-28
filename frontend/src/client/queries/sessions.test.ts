import {it, expect, vi} from 'vitest'
import {QueryClient, keepPreviousData} from '@tanstack/vue-query'
import {
	sessionsQuery,
	sessionKeys,
	deleteSessionMutationOptions,
} from './sessions'
const sdk = vi.hoisted(() => ({
	sessionsList: vi.fn(),
	sessionsDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	success: vi.fn(),
	error: vi.fn(),
}))

it('loads only the requested session page', async () => {
	const client = new QueryClient()
	sdk.sessionsList.mockResolvedValue({data: {
		items: [{id: 'one'}],
		page: 2,
		per_page: 25,
		total: 26,
		total_pages: 2,
	}})
	const options = sessionsQuery(2)
	expect(options.queryKey).toEqual(sessionKeys.list(2))
	expect(options.placeholderData).toBe(keepPreviousData)
	expect(options.staleTime).toBe(0)
	expect(await client.fetchQuery(options)).toEqual({
		items: [{id: 'one'}],
		page: 2,
		per_page: 25,
		total: 26,
		total_pages: 2,
	})
	expect(sdk.sessionsList).toHaveBeenCalledWith(expect.objectContaining({
		query: {
			page: 2,
			per_page: 25,
		},
	}))
	expect(sdk.sessionsList).toHaveBeenCalledTimes(1)
})

it('removes a revoked session and rewrites the totals on every cached page', async () => {
	const client = new QueryClient()
	client.setQueryData(sessionKeys.list(1), {
		items: [{id: 'first'}],
		page: 1,
		per_page: 25,
		total: 26,
		total_pages: 2,
	})
	client.setQueryData(sessionKeys.list(2), {
		items: [{id: 'revoke'}],
		page: 2,
		per_page: 25,
		total: 26,
		total_pages: 2,
	})
	sdk.sessionsDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteSessionMutationOptions()).execute('revoke')
	expect(sdk.sessionsDelete).toHaveBeenCalledWith({path: {session: 'revoke'}})
	expect(client.getQueryData(sessionKeys.list(1))).toEqual({
		items: [{id: 'first'}],
		page: 1,
		per_page: 25,
		total: 25,
		total_pages: 1,
	})
	expect(client.getQueryData(sessionKeys.list(2))).toEqual({
		items: [],
		page: 2,
		per_page: 25,
		total: 25,
		total_pages: 1,
	})
	expect(client.getQueryState(sessionKeys.list(1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(sessionKeys.list(2))?.isInvalidated).toBe(true)
})

it('leaves cached pages untouched when revoking fails', async () => {
	const client = new QueryClient()
	const cached = {
		items: [{id: 'revoke'}],
		page: 1,
		per_page: 25,
		total: 1,
		total_pages: 1,
	}
	client.setQueryData(sessionKeys.list(1), structuredClone(cached))
	sdk.sessionsDelete.mockRejectedValue(new Error('forbidden'))
	await expect(client.getMutationCache().build(client, deleteSessionMutationOptions()).execute('revoke'))
		.rejects.toThrow('forbidden')
	expect(client.getQueryData(sessionKeys.list(1))).toEqual(cached)
})
