import {it, expect, vi} from 'vitest'
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

it('uses separate operations and caches for user and project subscriptions', async () => {
	const client = new QueryClient()
	sdk.webhooksList.mockResolvedValue({data: {items: [{id: 1}], total_pages: 1}})
	sdk.userWebhooksList.mockResolvedValue({data: {items: [{id: 2}], total_pages: 1}})
	await client.fetchQuery(webhooksQuery({kind: 'project', projectId: 7}))
	await client.fetchQuery(webhooksQuery({kind: 'user'}))
	expect(sdk.webhooksList).toHaveBeenCalledWith({
		path: {project: 7},
		query: {page: 1},
		signal: expect.anything(),
	})
	expect(sdk.userWebhooksList).toHaveBeenCalledWith({
		query: {page: 1},
		signal: expect.anything(),
	})
	sdk.webhooksDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteWebhookMutationOptions()).execute({
		scope: {kind: 'project', projectId: 7},
		id: 1,
	})
	expect(sdk.webhooksDelete).toHaveBeenCalledWith({path: {project: 7, webhook: 1}})
	expect(client.getQueryData(webhookKeys.list({kind: 'project', projectId: 7}))).toEqual([])
	expect(client.getQueryState(webhookKeys.list({kind: 'project', projectId: 7}))?.isInvalidated).toBe(true)
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}))?.isInvalidated).toBe(false)
	expect(client.getQueryData(webhookKeys.list({kind: 'user'}))).toEqual([{id: 2}])
})
it('deletes user webhooks through the user operation and leaves project lists alone', async () => {
	const client = new QueryClient()
	client.setQueryData(webhookKeys.list({kind: 'user'}), [{id: 1}, {id: 2}])
	client.setQueryData(webhookKeys.list({kind: 'project', projectId: 7}), [{id: 1}])
	sdk.userWebhooksDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteWebhookMutationOptions()).execute({
		scope: {kind: 'user'},
		id: 1,
	})
	expect(sdk.userWebhooksDelete).toHaveBeenCalledWith({path: {webhook: 1}})
	expect(client.getQueryData(webhookKeys.list({kind: 'user'}))).toEqual([{id: 2}])
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}))?.isInvalidated).toBe(true)
	expect(client.getQueryData(webhookKeys.list({kind: 'project', projectId: 7}))).toEqual([{id: 1}])
	expect(client.getQueryState(webhookKeys.list({kind: 'project', projectId: 7}))?.isInvalidated).toBe(false)
})
it('creates through the scoped operation and stales only that scope', async () => {
	const client = new QueryClient()
	client.setQueryData(webhookKeys.list({kind: 'project', projectId: 7}), [])
	client.setQueryData(webhookKeys.list({kind: 'user'}), [])
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
	expect(client.getQueryState(webhookKeys.list({kind: 'project', projectId: 7}))?.isInvalidated).toBe(true)
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}))?.isInvalidated).toBe(false)
	await client.getMutationCache().build(client, createWebhookMutationOptions()).execute({
		scope: {kind: 'user'},
		body: {target_url: 'https://example.com/user'},
	})
	expect(sdk.userWebhooksCreate).toHaveBeenCalledWith({body: {target_url: 'https://example.com/user'}})
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}))?.isInvalidated).toBe(true)
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
