import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'
import type {IReactionPerEntity} from '@/modelTypes/IReaction'

export interface ITaskComment extends IAbstract {
	id: number
	taskId: ITask['id']
	comment: string
	author: IUser
	
	reactions: IReactionPerEntity

	created: Date
	updated: Date
}
