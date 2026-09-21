import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {webhookKeys, webhooksQuery, deleteWebhookMutationOptions} from './webhooks'
const sdk = vi.hoisted(() => ({
	webhooksList: vi.fn(),
	userWebhooksList: vi.fn(),
	webhooksDelete: vi.fn(),
	userWebhooksDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('uses separate operations and caches for user and project subscriptions', async () => {
	const client = new QueryClient()
	sdk.webhooksList.mockResolvedValue({data: {items: [{id: 1}], total_pages: 1}})
	sdk.userWebhooksList.mockResolvedValue({data: {items: [{id: 2}], total_pages: 1}})
	await client.fetchQuery(webhooksQuery({kind: 'project', projectId: 7}))
	await client.fetchQuery(webhooksQuery({kind: 'user'}))
	sdk.webhooksDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteWebhookMutationOptions()).execute({
		scope: {kind: 'project', projectId: 7},
		id: 1,
	})
	expect(sdk.webhooksDelete).toHaveBeenCalledWith({path: {project: 7, webhook: 1}})
	expect(client.getQueryData(webhookKeys.list({kind: 'project', projectId: 7}))).toEqual([])
	expect(client.getQueryState(webhookKeys.list({kind: 'user'}))?.isInvalidated).toBe(false)
	expect(client.getQueryData(webhookKeys.list({kind: 'user'}))).toEqual([{id: 2}])
})
