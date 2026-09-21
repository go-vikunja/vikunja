import AbstractService from '@/services/abstractService'
import type {BotUser as IUser} from '@/client/generated'
import type {IAbstract} from '@/modelTypes/IAbstract'

export default class BotUserService extends AbstractService<IUser & IAbstract> {
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
		return {id: 0, maxPermission: null, ...data}
	}
}
