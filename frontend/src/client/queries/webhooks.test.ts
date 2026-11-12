import {it, expect, vi, beforeEach} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	webhookKeys,
	webhooksQuery,
	webhookEventsQuery,
	createWebhookMutationOptions,
	deleteWebhookMutationOptions,
} from './webhooks'
const sdk = vi.hoisted(() => ({
	webhooksList: vi.fn(),
	userWebhooksList: vi.fn(),
	webhooksCreate: vi.fn(),
	userWebhooksCreate: vi.fn(),
	webhooksDelete: vi.fn(),
	userWebhooksDelete: vi.fn(),
	webhooksEventsList: vi.fn(),
	userWebhooksEvents: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
beforeEach(() => vi.clearAllMocks())

const PER_PAGE = 25

function webhookPage(items: {id: number}[], page = 1, total = items.length) {
	return {
		items,
		page,
		per_page: PER_PAGE,
		total,
		total_pages: Math.ceil(total / PER_PAGE),
	}
}

it('requests one page of the scoped list', async () => {
	const client = new QueryClient()
	sdk.webhooksList.mockResolvedValue({data: webhookPage([{id: 1}], 2, 26)})
	sdk.userWebhooksList.mockResolvedValue({data: {items: null}})
	expect(await client.fetchQuery(webhooksQuery({kind: 'project', projectId: 7}, 2))).toEqual(webhookPage([{id: 1}], 2, 26))
	expect(await client.fetchQuery(webhooksQuery({kind: 'user'}, 1))).toEqual({
		items: [],
		page: 1,
		per_page: 0,
		total: 0,
		total_pages: 0,
	})
	expect(sdk.webhooksList).toHaveBeenCalledTimes(1)
	expect(sdk.webhooksList).toHaveBeenCalledWith({
		path: {project: 7},
		query: {page: 2},
		signal: expect.anything(),
	})
	expect(sdk.userWebhooksList).toHaveBeenCalledTimes(1)
	expect(sdk.userWebhooksList).toHaveBeenCalledWith({
		query: {page: 1},
		signal: expect.anything(),
	})
})
it('removes a deleted project webhook from every cached page of that project only', async () => {
	const client = new QueryClient()
	const project = {kind: 'project', projectId: 7} as const
	const page1 = Array.from({length: PER_PAGE}, (_, i) => ({id: i + 1}))
	client.setQueryData(webhookKeys.list(project, 1), webhookPage(page1, 1, 26))
	client.setQueryData(webhookKeys.list(project, 2), webhookPage([{id: 26}], 2, 26))
	client.setQueryData(webhookKeys.list({kind: 'project', projectId: 8}, 1), webhookPage([{id: 26}]))
	client.setQueryData(webhookKeys.list({kind: 'user'}, 1), webhookPage([{id: 26}]))
	sdk.webhooksDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteWebhookMutationOptions()).execute({
		scope: project,
		id: 26,
	})
	expect(sdk.webhooksDelete).toHaveBeenCalledWith({path: {
		project: 7,
		webhook: 26,
	}})
	expect(client.getQueryData(webhookKeys.list(project, 1))).toEqual(webhookPage(page1, 1, 25))
	expect(client.getQueryData(webhookKeys.list(project, 2))).toEqual({
		...webhookPage([], 2, 25),
		total_pages: 1,
	})
	expect(client.getQueryState(webhookKeys.list(project, 1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(webhookKeys.list(project, 2))?.isInvalidated).toBe(true)
	expect(client.getQueryData(webhookKeys.list({kind: 'project', projectId: 8}, 1))).toEqual(webhookPage([{id: 26}]))
	expect(client.getQueryState(webhookKeys.list({kind: 'project', projectId: 8}, 1))?.isInvalidated).toBe(false)
	expect(client.getQueryData(webhookKeys.list({kind: 'user'}, 1))).toEqual(webhookPage([{id: 26}]))
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}, 1))?.isInvalidated).toBe(false)
})
it('deletes user webhooks through the user operation and leaves project lists alone', async () => {
	const client = new QueryClient()
	client.setQueryData(webhookKeys.list({kind: 'user'}, 1), webhookPage([{id: 1}, {id: 2}]))
	client.setQueryData(webhookKeys.list({kind: 'project', projectId: 7}, 1), webhookPage([{id: 1}]))
	sdk.userWebhooksDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteWebhookMutationOptions()).execute({
		scope: {kind: 'user'},
		id: 1,
	})
	expect(sdk.userWebhooksDelete).toHaveBeenCalledWith({path: {webhook: 1}})
	expect(client.getQueryData(webhookKeys.list({kind: 'user'}, 1))).toEqual(webhookPage([{id: 2}]))
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}, 1))?.isInvalidated).toBe(true)
	expect(client.getQueryData(webhookKeys.list({kind: 'project', projectId: 7}, 1))).toEqual(webhookPage([{id: 1}]))
	expect(client.getQueryState(webhookKeys.list({kind: 'project', projectId: 7}, 1))?.isInvalidated).toBe(false)
})
it('creates through the scoped operation and stales every page of only that scope', async () => {
	const client = new QueryClient()
	client.setQueryData(webhookKeys.list({kind: 'project', projectId: 7}, 1), webhookPage([]))
	client.setQueryData(webhookKeys.list({kind: 'project', projectId: 7}, 2), webhookPage([], 2))
	client.setQueryData(webhookKeys.list({kind: 'user'}, 1), webhookPage([]))
	sdk.webhooksCreate.mockResolvedValue({data: {id: 3}})
	sdk.userWebhooksCreate.mockResolvedValue({data: {id: 4}})
	const created = await client.getMutationCache().build(client, createWebhookMutationOptions()).execute({
		scope: {kind: 'project', projectId: 7},
		body: {target_url: 'https://example.com/project'},
	})
	expect(created).toEqual({id: 3})
	expect(sdk.webhooksCreate).toHaveBeenCalledWith({
		path: {project: 7},
		body: {target_url: 'https://example.com/project'},
	})
	expect(client.getQueryState(webhookKeys.list({kind: 'project', projectId: 7}, 1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(webhookKeys.list({kind: 'project', projectId: 7}, 2))?.isInvalidated).toBe(true)
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}, 1))?.isInvalidated).toBe(false)
	await client.getMutationCache().build(client, createWebhookMutationOptions()).execute({
		scope: {kind: 'user'},
		body: {target_url: 'https://example.com/user'},
	})
	expect(sdk.userWebhooksCreate).toHaveBeenCalledWith({body: {target_url: 'https://example.com/user'}})
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}, 1))?.isInvalidated).toBe(true)
})
it('reads the available events from the operation matching the scope', async () => {
	const client = new QueryClient()
	sdk.webhooksEventsList.mockResolvedValue({data: ['task.created']})
	sdk.userWebhooksEvents.mockResolvedValue({data: undefined})
	expect(await client.fetchQuery(webhookEventsQuery('project'))).toEqual(['task.created'])
	expect(await client.fetchQuery(webhookEventsQuery('user'))).toEqual([])
	expect(sdk.webhooksEventsList).toHaveBeenCalledWith({signal: expect.anything()})
	expect(sdk.userWebhooksEvents).toHaveBeenCalledWith({signal: expect.anything()})
})
