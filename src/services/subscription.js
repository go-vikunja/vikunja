import AbstractService from '@/services/abstractService'
import SubscriptionModel from '@/models/subscription'

export default class SubscriptionService extends AbstractService {
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