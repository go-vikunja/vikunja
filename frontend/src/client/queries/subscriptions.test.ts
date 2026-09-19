import {expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {normalizeTask, taskKeys} from './tasks'
import {setTaskSubscriptionMutationOptions} from './subscriptions'
const sdk = vi.hoisted(() => ({subscriptionsCreate: vi.fn(), subscriptionsDelete: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))

it('sets and clears a task subscription without creating unmounted queries', async () => {
	const client = new QueryClient()
	client.setQueryData(taskKeys.detail(1), normalizeTask({id: 1}))
	const subscription = {id: 3, entity: 'task', entity_id: 1}
	sdk.subscriptionsCreate.mockResolvedValue({data: subscription})
	await client.getMutationCache().build(client, setTaskSubscriptionMutationOptions())
		.execute({taskId: 1, subscribed: true})
	expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({subscription})
	expect(sdk.subscriptionsCreate).toHaveBeenCalledWith({path: {entity: 'task', entityID: 1}})
	sdk.subscriptionsDelete.mockResolvedValue({data: undefined})
	await client.getMutationCache().build(client, setTaskSubscriptionMutationOptions())
		.execute({taskId: 1, subscribed: false})
	expect(client.getQueryData(taskKeys.detail(1))).toHaveProperty('subscription', undefined)
	expect(client.getQueryCache().getAll()).toHaveLength(1)
})
