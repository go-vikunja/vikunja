import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {apiTokenKeys, apiTokensQuery, createApiTokenMutationOptions, deleteApiTokenMutationOptions} from './apiTokens'
const sdk = vi.hoisted(() => ({tokensList: vi.fn(), tokensCreate: vi.fn(), tokensDelete: vi.fn(), tokenRoutes: vi.fn(), mcpInfo: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('loads all pages with the bot owner filter', async () => {
	const client = new QueryClient()
	sdk.tokensList.mockImplementation(({query}) => Promise.resolve({data: {items: [{id: query.page}], total_pages: 2}}))
	expect(await client.fetchQuery(apiTokensQuery(5))).toEqual([{id: 1}, {id: 2}])
	expect(sdk.tokensList).toHaveBeenLastCalledWith(expect.objectContaining({query: {page: 2, owner_id: 5}}))
})
it('never stores a newly created plaintext token in a list cache', async () => {
	const client = new QueryClient()
	client.setQueryData(apiTokenKeys.list(0), [{id: 1}])
	sdk.tokensCreate.mockResolvedValue({data: {id: 2, token: 'tk_secret'}})
	const result = await client.getMutationCache().build(client, createApiTokenMutationOptions()).execute({title: 'New'})
	expect(result.token).toBe('tk_secret')
	expect(client.getQueryData(apiTokenKeys.list(0))).toEqual([{id: 1}])
	expect(client.getQueryState(apiTokenKeys.list(0))?.isInvalidated).toBe(true)
})
it('removes revoked tokens from all existing owner lists', async () => {
	const client = new QueryClient()
	client.setQueryData(apiTokenKeys.list(0), [{id: 1}, {id: 2}])
	client.setQueryData(apiTokenKeys.list(5), [{id: 2}])
	sdk.tokensDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteApiTokenMutationOptions()).execute(2)
	expect(client.getQueryData(apiTokenKeys.list(0))).toEqual([{id: 1}])
	expect(client.getQueryData(apiTokenKeys.list(5))).toEqual([])
})
