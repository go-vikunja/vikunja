import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {sessionsQuery, sessionKeys, deleteSessionMutationOptions} from './sessions'
const sdk = vi.hoisted(() => ({sessionsList: vi.fn(), sessionsDelete: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('loads the first session page through the generated operation', async () => {
	const client = new QueryClient()
	sdk.sessionsList.mockResolvedValue({data: {items: [{id: 'one'}], total_pages: 1}})
	expect(await client.fetchQuery(sessionsQuery())).toEqual([{id: 'one'}])
	expect(sdk.sessionsList).toHaveBeenCalledWith(expect.objectContaining({query: {page: 1}}))
})
it('removes a revoked session from an existing cache and marks the list stale', async () => {
	const client = new QueryClient()
	client.setQueryData(sessionKeys.all, [{id: 'keep'}, {id: 'revoke'}])
	sdk.sessionsDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteSessionMutationOptions()).execute('revoke')
	expect(sdk.sessionsDelete).toHaveBeenCalledWith({path: {session: 'revoke'}})
	expect(client.getQueryData(sessionKeys.all)).toEqual([{id: 'keep'}])
	expect(client.getQueryState(sessionKeys.all)?.isInvalidated).toBe(true)
})
