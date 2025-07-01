import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IUser} from '@/modelTypes/IUser'
import ReactionModel from '@/models/reaction'
import UserModel from '@/models/user'

export default class ReactionService extends AbstractService {
	constructor() {
		super({
			getAll: '{kind}/{id}/reactions',
			create: '{kind}/{id}/reactions',
			delete: '{kind}/{id}/reactions/delete',
		})
	}

	modelFactory(data: Partial<IAbstract>): ReactionModel {
		return new ReactionModel(data)
	}

	modelGetAllFactory(data: Record<string, Partial<IUser>[]>): Record<string, IUser[]> {
		Object.keys(data).forEach(reaction => {
			data[reaction] = data[reaction]?.map((u: Partial<IUser>) => new UserModel(u))
		})

		return data as Record<string, IUser[]>
	}

	async delete(model: IAbstract) {
		const finalUrl = this.getReplacedRoute(this.paths.delete, model as Record<string, unknown>)
		return super.post(finalUrl, model)
	}
}
