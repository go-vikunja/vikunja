import {
	subscriptionsCreate,
	subscriptionsDelete,
} from '@/client/generated'
import type {Subscription} from '@/client/generated'

export async function setSubscription(
	entity: 'task' | 'project',
	entityID: number,
	subscribed: boolean,
): Promise<Subscription | undefined> {
	const path = {entity, entityID} as const
	if (!subscribed) {
		await subscriptionsDelete({path})
		return undefined
	}
	return (await subscriptionsCreate({path})).data
}
