import AbstractService from '@/services/abstractService'
import type {IUser} from '@/modelTypes/IUser'
import UserModel from '@/models/user'

export default class BotUserService extends AbstractService<IUser> {
	constructor() {
		super({
			create: '/user/bots',
			getAll: '/user/bots',
			get: '/user/bots/{id}',
			update: '/user/bots/{id}',
			delete: '/user/bots/{id}',
		})
	}

	modelFactory(data: Partial<IUser>) {
		return new UserModel(data)
	}
}
