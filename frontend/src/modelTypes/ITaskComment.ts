import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'

export interface ITaskComment extends IAbstract {
	id: number
	taskId: ITask['id']
	comment: string
	author: IUser

	created: Date
	updated: Date
}