import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import ReactionModel from '@/models/reaction'
import type {IReactionPerEntity, IReaction} from '@/modelTypes/IReaction'
import UserModel from '@/models/user'

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

	modelGetAllFactory(data: Partial<IReactionPerEntity & { maxPermission?: unknown }>): IReactionPerEntity & { maxPermission?: unknown } {
		const result: IReactionPerEntity & { maxPermission?: unknown } = {}

		Object.keys(data).forEach(reaction => {
			if (reaction !== 'maxPermission') {
				result[reaction] = data[reaction]?.map((u) => new UserModel(u)) || []
			}
		})

		// Preserve maxPermission if it exists
		if ('maxPermission' in data) {
			result.maxPermission = data.maxPermission
		}

		return result
	}

	async delete(model: IReaction) {
		const finalUrl = this.getReplacedRoute(this.paths.delete, model)
		return super.post(finalUrl, model)
	}
}
