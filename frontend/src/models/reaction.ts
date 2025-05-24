import type {IReaction} from '@/modelTypes/IReaction'
import AbstractModel from '@/models/abstractModel'

export default class ReactionModel extends AbstractModel<IReaction> implements IReaction {
	id: number = 0
	kind: 'tasks' | 'comments' = 'tasks'
	value: string = ''
	
	constructor(data: Partial<IReaction>) {
		super()
		this.assignData(data)
	}
}
	
