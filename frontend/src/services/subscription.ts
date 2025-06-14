import AbstractService from '@/services/abstractService'
import SubscriptionModel from '@/models/subscription'
import type {ISubscription} from '@/modelTypes/ISubscription'

export default class SubscriptionService extends AbstractService<ISubscription> {
	constructor() {
		super({
			create: '/subscriptions/{entity}/{entityId}',
			delete: '/subscriptions/{entity}/{entityId}',
		})
	}

	modelFactory(data) {
		return new SubscriptionModel(data)
	}
}
