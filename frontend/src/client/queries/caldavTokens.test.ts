import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {caldavTokenKeys, caldavTokensQuery, createCaldavTokenMutationOptions, deleteCaldavTokenMutationOptions} from './caldavTokens'
const sdk = vi.hoisted(() => ({
	caldavTokensCreate: vi.fn(),
	caldavTokensList: vi.fn(),
	caldavTokensDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('returns the one-time secret without putting it in the shared list cache', async () => {
	const client = new QueryClient()
	client.setQueryData(caldavTokenKeys.all, [{id: 1}])
	sdk.caldavTokensCreate.mockResolvedValue({data: {id: 2, token: 'one-time-secret'}})
	const created = await client.getMutationCache().build(client, createCaldavTokenMutationOptions()).execute(undefined)
	expect(created.token).toBe('one-time-secret')
	expect(client.getQueryData(caldavTokenKeys.all)).toEqual([{id: 1}])
	expect(client.getQueryState(caldavTokenKeys.all)?.isInvalidated).toBe(true)
})

it('does not keep the one-time secret in the mutation cache', async () => {
	const client = new QueryClient()
	sdk.caldavTokensCreate.mockResolvedValue({data: {id: 2, token: 'one-time-secret'}})
	await client.getMutationCache().build(client, createCaldavTokenMutationOptions()).execute(undefined)
	await new Promise(resolve => setTimeout(resolve))
	expect(client.getMutationCache().getAll()).toEqual([])
})

it('loads the first token page through the generated operation', async () => {
	const client = new QueryClient()
	sdk.caldavTokensList.mockResolvedValue({data: {items: [{id: 1}], total_pages: 1}})
	expect(await client.fetchQuery(caldavTokensQuery())).toEqual([{id: 1}])
	expect(sdk.caldavTokensList).toHaveBeenCalledWith(expect.objectContaining({query: {page: 1}}))
})

it('removes a deleted token from an existing cache and marks the list stale', async () => {
	const client = new QueryClient()
	client.setQueryData(caldavTokenKeys.all, [{id: 1}, {id: 2}])
	sdk.caldavTokensDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteCaldavTokenMutationOptions()).execute(2)
	expect(sdk.caldavTokensDelete).toHaveBeenCalledWith({path: {id: 2}})
	expect(client.getQueryData(caldavTokenKeys.all)).toEqual([{id: 1}])
	expect(client.getQueryState(caldavTokenKeys.all)?.isInvalidated).toBe(true)
})
