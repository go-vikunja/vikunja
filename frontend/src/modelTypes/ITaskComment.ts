import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {Task as ITask} from '@/client/generated'
import type {TaskComment} from '@/client/generated'

export interface ITaskComment extends IAbstract {
	id: number
	taskId: ITask['id']
	comment: string
	author: IUser
	
	reactions: NonNullable<TaskComment['reactions']>

	created: Date
	updated: Date
}
