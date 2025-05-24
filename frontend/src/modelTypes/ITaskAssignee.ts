import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'
import type {IUser} from './IUser'

export interface ITaskAssignee extends IAbstract {
	created: Date
	userId: IUser['id']
	taskId: ITask['id']
}
