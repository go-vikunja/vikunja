import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
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

	modelGetAllFactory(data: any): any {
		Object.keys(data).forEach(reaction => {
			data[reaction] = data[reaction]?.map((u: any) => new UserModel(u))
		})

		return data
	}

	async delete(model: IAbstract) {
		const finalUrl = this.getReplacedRoute(this.paths.delete, model as any)
		return super.post(finalUrl, model)
	}
}
