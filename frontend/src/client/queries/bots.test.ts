import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	botKeys,
	botsQuery,
	createBotMutationOptions,
	updateBotMutationOptions,
	deleteBotMutationOptions,
} from './bots'
import {apiTokenKeys} from './apiTokens'
import type {BotUser} from '@/client/generated'
const sdk = vi.hoisted(() => ({
	botsList: vi.fn(),
	botsCreate: vi.fn(),
	botsUpdate: vi.fn(),
	botsDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	success: vi.fn(),
	error: vi.fn(),
}))

function page(pageNumber: number, items: BotUser[], total = 3) {
	return {
		items,
		page: pageNumber,
		per_page: 2,
		total,
		total_pages: Math.ceil(total / 2),
	}
}

it('loads one page of bots with a fixed page size', async () => {
	const client = new QueryClient()
	sdk.botsList.mockResolvedValue({data: {
		items: [{id: 21}],
		page: 2,
		per_page: 20,
		total: 21,
		total_pages: 2,
	}})
	expect(await client.fetchQuery(botsQuery(2))).toEqual({
		items: [{id: 21}],
		page: 2,
		per_page: 20,
		total: 21,
		total_pages: 2,
	})
	expect(sdk.botsList).toHaveBeenCalledOnce()
	expect(sdk.botsList).toHaveBeenCalledWith(expect.objectContaining({query: {
		page: 2,
		per_page: 20,
	}}))
	expect(client.getQueryData(botKeys.list(2))).toBeDefined()
})
it('creating a bot stales every cached page', async () => {
	const client = new QueryClient()
	client.setQueryData(botKeys.list(1), page(1, [{id: 1}, {id: 2}]))
	client.setQueryData(botKeys.list(2), page(2, [{id: 3}]))
	sdk.botsCreate.mockResolvedValue({data: {id: 4}})
	await client.getMutationCache().build(client, createBotMutationOptions()).execute({username: 'bot-new'})
	expect(client.getQueryState(botKeys.list(1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(botKeys.list(2))?.isInvalidated).toBe(true)
})
it('reconciles updated bot records on whichever page holds them and stales the list', async () => {
	const client = new QueryClient()
	client.setQueryData(botKeys.list(1), page(1, [{id: 1}, {id: 2}]))
	client.setQueryData(botKeys.list(2), page(2, [{id: 3, name: 'Old'}]))
	sdk.botsUpdate.mockResolvedValue({data: {
		id: 3,
		name: 'New',
		status: 2,
	}})
	await client.getMutationCache().build(client, updateBotMutationOptions()).execute({
		id: 3,
		body: {
			name: 'New',
			status: 2,
		},
	})
	expect(client.getQueryData(botKeys.list(1))).toEqual(page(1, [{id: 1}, {id: 2}]))
	expect(client.getQueryData(botKeys.list(2))).toEqual(page(2, [{
		id: 3,
		name: 'New',
		status: 2,
	}]))
	expect(client.getQueryState(botKeys.list(1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(botKeys.list(2))?.isInvalidated).toBe(true)
})
it('deleting a bot rewrites the totals of every cached page', async () => {
	const client = new QueryClient()
	client.setQueryData(botKeys.list(1), page(1, [{id: 1}, {id: 2}]))
	client.setQueryData(botKeys.list(2), page(2, [{id: 3}]))
	sdk.botsDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteBotMutationOptions()).execute(3)
	expect(client.getQueryData(botKeys.list(1))).toEqual({
		...page(1, [{id: 1}, {id: 2}]),
		total: 2,
		total_pages: 1,
	})
	expect(client.getQueryData(botKeys.list(2))).toEqual({
		...page(2, []),
		total: 2,
		total_pages: 1,
	})
	expect(client.getQueryState(botKeys.list(1))?.isInvalidated).toBe(true)
})
it('deleting a bot evicts its token cache and leaves other owners intact', async () => {
	const client = new QueryClient()
	client.setQueryData(botKeys.list(1), page(1, [{id: 1}, {id: 2}], 2))
	client.setQueryData(apiTokenKeys.list(1), [{id: 9}])
	client.setQueryData(apiTokenKeys.list(2), [{id: 10}])
	sdk.botsDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteBotMutationOptions()).execute(1)
	expect(client.getQueryData(botKeys.list(1))).toEqual(page(1, [{id: 2}], 1))
	expect(client.getQueryData(apiTokenKeys.list(1))).toBeUndefined()
	expect(client.getQueryData(apiTokenKeys.list(2))).toEqual([{id: 10}])
})
