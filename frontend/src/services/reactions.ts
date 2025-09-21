import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import ReactionModel from '@/models/reaction'
import type {IReaction} from '@/modelTypes/IReaction'

export default class ReactionService extends AbstractService<IReaction> {
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

	modelGetAllFactory(data: Partial<IReaction>): IReaction {
		return this.modelFactory(data)
	}

	async delete(model: IReaction) {
		const finalUrl = this.getReplacedRoute(this.paths.delete, model)
		const response = await super.post(finalUrl, model)
		return this.modelFactory(response)
	}
}
