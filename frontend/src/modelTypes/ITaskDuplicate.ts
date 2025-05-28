import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'

export interface ITaskDuplicate extends IAbstract {
	taskId: number
	projectId: number
	duplicatedTask: ITask | null
}