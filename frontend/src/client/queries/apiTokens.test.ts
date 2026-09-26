import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	apiTokenKeys,
	apiTokensQuery,
	botApiTokensQuery,
	createApiTokenMutationOptions,
	deleteApiTokenMutationOptions,
} from './apiTokens'
const sdk = vi.hoisted(() => ({
	tokensList: vi.fn(),
	tokensCreate: vi.fn(),
	tokensDelete: vi.fn(),
	tokenRoutes: vi.fn(),
	mcpInfo: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('loads all pages with the bot owner filter', async () => {
	const client = new QueryClient()
	sdk.tokensList.mockImplementation(({query}) => Promise.resolve({data: {items: [{id: query.page}], total_pages: 2}}))
	expect(await client.fetchQuery(botApiTokensQuery(5))).toEqual([{id: 1}, {id: 2}])
	expect(sdk.tokensList).toHaveBeenLastCalledWith(expect.objectContaining({query: {page: 2, owner_id: 5}}))
	expect(botApiTokensQuery(5).queryKey).toEqual(['apiTokens', 'list', 5])
})
it('omits the owner filter for the signed-in user', async () => {
	const client = new QueryClient()
	sdk.tokensList.mockResolvedValue({data: {items: [{id: 1}], total_pages: 1}})
	expect(await client.fetchQuery(apiTokensQuery())).toEqual([{id: 1}])
	expect(sdk.tokensList.mock.lastCall?.[0].query).toEqual({page: 1})
	expect(apiTokensQuery().queryKey).toEqual(['apiTokens', 'list', 'self'])
})
it('never stores a newly created plaintext token in a list cache', async () => {
	const client = new QueryClient()
	client.setQueryData(apiTokenKeys.ownList, [{id: 1}])
	sdk.tokensCreate.mockResolvedValue({data: {id: 2, token: 'tk_secret'}})
	const result = await client.getMutationCache().build(client, createApiTokenMutationOptions()).execute({title: 'New'})
	expect(result.token).toBe('tk_secret')
	expect(client.getQueryData(apiTokenKeys.ownList)).toEqual([{id: 1}])
	expect(client.getQueryState(apiTokenKeys.ownList)?.isInvalidated).toBe(true)
})
it('invalidates only the bot list when a token is created for a bot', async () => {
	const client = new QueryClient()
	client.setQueryData(apiTokenKeys.ownList, [{id: 1}])
	client.setQueryData(apiTokenKeys.list(5), [{id: 2}])
	sdk.tokensCreate.mockResolvedValue({data: {id: 3, token: 'tk_secret'}})
	await client.getMutationCache().build(client, createApiTokenMutationOptions()).execute({
		title: 'New',
		owner_id: 5,
	})
	expect(client.getQueryState(apiTokenKeys.list(5))?.isInvalidated).toBe(true)
	expect(client.getQueryState(apiTokenKeys.ownList)?.isInvalidated).toBe(false)
})
it('removes revoked tokens from all existing owner lists', async () => {
	const client = new QueryClient()
	client.setQueryData(apiTokenKeys.ownList, [{id: 1}, {id: 2}])
	client.setQueryData(apiTokenKeys.list(5), [{id: 2}])
	sdk.tokensDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteApiTokenMutationOptions()).execute(2)
	expect(client.getQueryData(apiTokenKeys.ownList)).toEqual([{id: 1}])
	expect(client.getQueryData(apiTokenKeys.list(5))).toEqual([])
})
