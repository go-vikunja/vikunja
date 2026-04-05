import AbstractService from '@/services/abstractService'
import type {IBotUser} from '@/modelTypes/IBotUser'
import BotUserModel from '@/models/botUser'

export default class BotUserService extends AbstractService<IBotUser> {
	constructor() {
		super({
			create: '/user/bots',
			getAll: '/user/bots',
			get: '/user/bots/{id}',
			update: '/user/bots/{id}',
			delete: '/user/bots/{id}',
		})
	}

	modelFactory(data: Partial<IBotUser>) {
		return new BotUserModel(data)
	}
}
