import {beforeEach, expect, it, vi} from 'vitest'
import {setSubscription} from './subscriptionRequests'

const sdk = vi.hoisted(() => ({
	subscriptionsCreate: vi.fn(),
	subscriptionsDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)

const subscription = {
	id: 3,
	entity: 'task',
	entity_id: 1,
} as const

beforeEach(() => {
	vi.clearAllMocks()
})

it('creates a subscription and returns its data', async () => {
	sdk.subscriptionsCreate.mockResolvedValue({data: subscription})
	await expect(setSubscription('task', 1, true)).resolves.toEqual(subscription)
	expect(sdk.subscriptionsCreate).toHaveBeenCalledWith({path: {
		entity: 'task',
		entityID: 1,
	}})
	expect(sdk.subscriptionsDelete).not.toHaveBeenCalled()
})

it('deletes a subscription and returns undefined', async () => {
	sdk.subscriptionsDelete.mockResolvedValue({data: undefined})
	await expect(setSubscription('project', 2, false)).resolves.toBeUndefined()
	expect(sdk.subscriptionsDelete).toHaveBeenCalledWith({path: {
		entity: 'project',
		entityID: 2,
	}})
	expect(sdk.subscriptionsCreate).not.toHaveBeenCalled()
})
