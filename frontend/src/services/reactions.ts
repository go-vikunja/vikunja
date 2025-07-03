import AbstractService from '@/services/abstractService'
import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IUser} from '@/modelTypes/IUser'
import type {IReactionPerEntity} from '@/modelTypes/IReaction'
import ReactionModel from '@/models/reaction'
import UserModel from '@/models/user'

export default class ReactionService extends AbstractService<IAbstract> {
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

	// Override with different signature for reaction-specific data structure
	modelGetAllFactory(data: Record<string, Partial<IUser>[]>): IReactionPerEntity
	modelGetAllFactory(data: Partial<IAbstract>): IAbstract
	modelGetAllFactory(data: Record<string, Partial<IUser>[]> | Partial<IAbstract>): IReactionPerEntity | IAbstract {
		// Handle reactions data structure
		if (data && typeof data === 'object' && !Array.isArray(data) && 'id' in data) {
			// Standard IAbstract data - delegate to base implementation
			return this.modelFactory(data as Partial<IAbstract>)
		}
		
		// Handle reactions data structure
		const reactionData = data as Record<string, Partial<IUser>[]>
		Object.keys(reactionData).forEach(reaction => {
			reactionData[reaction] = reactionData[reaction]?.map((u: Partial<IUser>) => new UserModel(u))
		})

		return reactionData as IReactionPerEntity
	}

	async delete(model: IAbstract) {
		const finalUrl = this.getReplacedRoute(this.paths.delete, model as unknown as Record<string, unknown>)
		return super.post(finalUrl, model)
	}
}
