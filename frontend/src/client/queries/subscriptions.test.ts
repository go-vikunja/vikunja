import {
	beforeEach,
	expect,
	it,
	vi,
} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	error,
	success,
} from '@/message'
import {
	normalizeTask,
	taskKeys,
} from './tasks'
import {setTaskSubscriptionMutationOptions} from './subscriptions'

const sdk = vi.hoisted(() => ({
	subscriptionsCreate: vi.fn(),
	subscriptionsDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

const subscription = {
	id: 3,
	entity: 'task',
	entity_id: 1,
} as const

let client: QueryClient
beforeEach(() => {
	vi.clearAllMocks()
	client = new QueryClient()
})

it('sets and clears a task subscription', async () => {
	client.setQueryData(taskKeys.detail(1), normalizeTask({id: 1}))
	sdk.subscriptionsCreate.mockResolvedValue({data: subscription})
	await client.getMutationCache().build(client, setTaskSubscriptionMutationOptions())
		.execute({
			taskId: 1,
			subscribed: true,
		})
	expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({subscription})
	expect(sdk.subscriptionsCreate).toHaveBeenCalledWith({path: {
		entity: 'task',
		entityID: 1,
	}})
	expect(success).toHaveBeenCalledWith({message: 'You are now subscribed to this task'})

	sdk.subscriptionsDelete.mockResolvedValue({data: undefined})
	await client.getMutationCache().build(client, setTaskSubscriptionMutationOptions())
		.execute({
			taskId: 1,
			subscribed: false,
		})
	expect(client.getQueryData(taskKeys.detail(1))).toHaveProperty('subscription', undefined)
	expect(sdk.subscriptionsDelete).toHaveBeenCalledWith({path: {
		entity: 'task',
		entityID: 1,
	}})
	expect(success).toHaveBeenLastCalledWith({message: 'You are now unsubscribed to this task'})
})

it('does not create a cache entry for an unrelated task id', async () => {
	client.setQueryData(taskKeys.detail(1), normalizeTask({id: 1}))
	const entryCountBefore = client.getQueryCache().getAll().length
	sdk.subscriptionsCreate.mockResolvedValue({data: subscription})
	await client.getMutationCache().build(client, setTaskSubscriptionMutationOptions())
		.execute({
			taskId: 99,
			subscribed: true,
		})
	expect(client.getQueryData(taskKeys.detail(99))).toBeUndefined()
	expect(client.getQueryCache().getAll().length).toBe(entryCountBefore)
})

it('invalidates every cached detail expansion but leaves task lists alone', async () => {
	const detailKey = taskKeys.detail(1, ['reactions'])
	client.setQueryData(detailKey, normalizeTask({id: 1}))
	const listKey = taskKeys.list({project: 1})
	client.setQueryData(listKey, {
		items: [normalizeTask({id: 1})],
		total: 1,
		per_page: 50,
		total_pages: 1,
	})
	sdk.subscriptionsCreate.mockResolvedValue({data: subscription})
	await client.getMutationCache().build(client, setTaskSubscriptionMutationOptions())
		.execute({
			taskId: 1,
			subscribed: true,
		})
	expect(client.getQueryState(detailKey)?.isInvalidated).toBe(true)
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(false)
})

it('keeps a failed subscription change out of the cache', async () => {
	client.setQueryData(taskKeys.detail(1), normalizeTask({
		id: 1,
		subscription,
	}))
	sdk.subscriptionsDelete.mockRejectedValue(new Error('nope'))
	await expect(client.getMutationCache().build(client, setTaskSubscriptionMutationOptions())
		.execute({
			taskId: 1,
			subscribed: false,
		})).rejects.toThrow('nope')
	expect(client.getQueryData(taskKeys.detail(1))).toMatchObject({subscription})
	expect(error).toHaveBeenCalledTimes(1)
	expect(success).not.toHaveBeenCalled()
})
