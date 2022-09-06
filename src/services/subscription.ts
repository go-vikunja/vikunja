import AbstractService from '@/services/abstractService'
import SubscriptionModel, { type ISubscription } from '@/models/subscription'

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