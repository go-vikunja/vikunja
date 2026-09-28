import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	apiTokenKeys,
	apiTokensQuery,
	allApiTokensQuery,
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

function page(items: {id: number}[], page: number, total: number, perPage = 25) {
	return {
		items,
		page,
		per_page: perPage,
		total,
		total_pages: Math.ceil(total / perPage),
	}
}

it('requests one page of the signed-in user\'s tokens', async () => {
	const client = new QueryClient()
	sdk.tokensList.mockResolvedValue({data: page([{id: 26}], 2, 26)})
	expect(await client.fetchQuery(apiTokensQuery(2))).toEqual(page([{id: 26}], 2, 26))
	expect(sdk.tokensList).toHaveBeenCalledOnce()
	expect(sdk.tokensList.mock.lastCall?.[0].query).toEqual({
		page: 2,
		per_page: 25,
	})
	expect(apiTokensQuery(2).queryKey).toEqual(['apiTokens', 'list', 'self', 2])
})
it('requests one page of a bot\'s tokens with the owner filter', async () => {
	const client = new QueryClient()
	sdk.tokensList.mockResolvedValue({data: page([{id: 1}], 1, 1, 50)})
	expect(await client.fetchQuery(botApiTokensQuery(5, 1))).toEqual(page([{id: 1}], 1, 1, 50))
	expect(sdk.tokensList).toHaveBeenLastCalledWith(expect.objectContaining({query: {
		page: 1,
		per_page: 50,
		owner_id: 5,
	}}))
	expect(botApiTokensQuery(5, 1).queryKey).toEqual(['apiTokens', 'list', 5, 1])
})
it('does not show another bot\'s page as placeholder', () => {
	const placeholder = botApiTokensQuery(5, 1).placeholderData as (data: unknown, query: unknown) => unknown
	const previous = page([{id: 1}], 2, 51, 50)
	expect(placeholder(previous, {queryKey: apiTokenKeys.page(5, 2)})).toBe(previous)
	expect(placeholder(previous, {queryKey: apiTokenKeys.page(6, 1)})).toBeUndefined()
})
it('loads every page of the signed-in user\'s tokens for the MCP filter', async () => {
	const client = new QueryClient()
	sdk.tokensList.mockImplementation(({query}) => Promise.resolve({data: {
		items: [{id: query.page}],
		total_pages: 2,
	}}))
	expect(await client.fetchQuery(allApiTokensQuery())).toEqual([{id: 1}, {id: 2}])
	expect(sdk.tokensList.mock.lastCall?.[0].query).toEqual({page: 2})
	expect(allApiTokensQuery().queryKey).toEqual(['apiTokens', 'all'])
})
it('never stores a newly created plaintext token in a list cache', async () => {
	const client = new QueryClient()
	client.setQueryData(apiTokenKeys.ownPage(1), page([{id: 1}], 1, 1))
	client.setQueryData(apiTokenKeys.ownAll, [{id: 1}])
	sdk.tokensCreate.mockResolvedValue({data: {id: 2, token: 'tk_secret'}})
	const result = await client.getMutationCache().build(client, createApiTokenMutationOptions()).execute({title: 'New'})
	expect(result.token).toBe('tk_secret')
	expect(client.getQueryData(apiTokenKeys.ownPage(1))).toEqual(page([{id: 1}], 1, 1))
	expect(client.getQueryData(apiTokenKeys.ownAll)).toEqual([{id: 1}])
	expect(client.getQueryState(apiTokenKeys.ownPage(1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(apiTokenKeys.ownAll)?.isInvalidated).toBe(true)
})
it('invalidates only the bot list when a token is created for a bot', async () => {
	const client = new QueryClient()
	client.setQueryData(apiTokenKeys.ownPage(1), page([{id: 1}], 1, 1))
	client.setQueryData(apiTokenKeys.ownAll, [{id: 1}])
	client.setQueryData(apiTokenKeys.page(5, 1), page([{id: 2}], 1, 1, 50))
	sdk.tokensCreate.mockResolvedValue({data: {id: 3, token: 'tk_secret'}})
	await client.getMutationCache().build(client, createApiTokenMutationOptions()).execute({
		title: 'New',
		owner_id: 5,
	})
	expect(client.getQueryState(apiTokenKeys.page(5, 1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(apiTokenKeys.ownPage(1))?.isInvalidated).toBe(false)
	expect(client.getQueryState(apiTokenKeys.ownAll)?.isInvalidated).toBe(false)
})
it('removes a revoked token and rewrites the totals of every cached page of its list', async () => {
	const client = new QueryClient()
	const firstPage = Array.from({length: 25}, (_, index) => ({id: index + 1}))
	client.setQueryData(apiTokenKeys.ownPage(1), page(firstPage, 1, 26))
	client.setQueryData(apiTokenKeys.ownPage(2), page([{id: 26}], 2, 26))
	client.setQueryData(apiTokenKeys.ownAll, [...firstPage, {id: 26}])
	client.setQueryData(apiTokenKeys.page(5, 1), page([{id: 30}], 1, 1, 50))
	sdk.tokensDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteApiTokenMutationOptions()).execute(26)
	expect(client.getQueryData(apiTokenKeys.ownPage(1))).toEqual(page(firstPage, 1, 25))
	expect(client.getQueryData(apiTokenKeys.ownPage(2))).toEqual(page([], 2, 25))
	expect(client.getQueryData(apiTokenKeys.ownAll)).toEqual(firstPage)
	expect(client.getQueryData(apiTokenKeys.page(5, 1))).toEqual(page([{id: 30}], 1, 1, 50))
	expect(client.getQueryState(apiTokenKeys.ownPage(1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(apiTokenKeys.page(5, 1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(apiTokenKeys.ownAll)?.isInvalidated).toBe(true)
})
