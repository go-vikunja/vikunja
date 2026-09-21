import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {botKeys, updateBotMutationOptions, deleteBotMutationOptions} from './bots'
import {apiTokenKeys} from './apiTokens'
const sdk = vi.hoisted(() => ({botsList: vi.fn(), botsCreate: vi.fn(), botsUpdate: vi.fn(), botsDelete: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('reconciles updated bot records and stales the list', async () => {
	const client = new QueryClient()
	client.setQueryData(botKeys.all, [{id: 1, name: 'Old'}, {id: 2}])
	sdk.botsUpdate.mockResolvedValue({data: {id: 1, name: 'New', status: 2}})
	await client.getMutationCache().build(client, updateBotMutationOptions()).execute({id: 1, body: {name: 'New', status: 2}})
	expect(client.getQueryData(botKeys.all)).toEqual([{id: 1, name: 'New', status: 2}, {id: 2}])
	expect(client.getQueryState(botKeys.all)?.isInvalidated).toBe(true)
})
it('deleting a bot evicts its token cache and leaves other owners intact', async () => {
	const client = new QueryClient()
	client.setQueryData(botKeys.all, [{id: 1}, {id: 2}])
	client.setQueryData(apiTokenKeys.list(1), [{id: 9}])
	client.setQueryData(apiTokenKeys.list(2), [{id: 10}])
	sdk.botsDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteBotMutationOptions()).execute(1)
	expect(client.getQueryData(botKeys.all)).toEqual([{id: 2}])
	expect(client.getQueryData(apiTokenKeys.list(1))).toBeUndefined()
	expect(client.getQueryData(apiTokenKeys.list(2))).toEqual([{id: 10}])
})
