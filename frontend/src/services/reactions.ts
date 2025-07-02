import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IUser} from '@/modelTypes/IUser'
import ReactionModel from '@/models/reaction'
import UserModel from '@/models/user'

export default class ReactionService extends AbstractService<ReactionModel> {
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

	// Special handling for reactions getAll - returns raw reaction data
	async getAll(model: ReactionModel, params: Record<string, unknown> = {}, page = 1): Promise<Record<string, IUser[]>> {
		const cancel = this.setLoading()
		model = this.beforeGet(model)
		const finalUrl = this.getReplacedRoute(this.paths.getAll, model as Record<string, unknown>)

		try {
			const response = await this.http.get(finalUrl, {params})
			const data = response.data as Record<string, Partial<IUser>[]>
			
			const processedReactions: Record<string, IUser[]> = {}
			Object.keys(data).forEach(reaction => {
				processedReactions[reaction] = (data[reaction] || []).map((u: Partial<IUser>) => new UserModel(u))
			})

			return processedReactions
		} finally {
			cancel()
		}
	}

	async delete(model: IAbstract) {
		const finalUrl = this.getReplacedRoute(this.paths.delete, model as Record<string, unknown>)
		return super.post(finalUrl, model as ReactionModel)
	}
}
